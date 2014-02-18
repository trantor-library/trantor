package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

func GetTags(numTags int, tagsColl *mgo.Collection) ([]string, error) {
	var result []struct {
		Tag string "_id"
	}
	err := tagsColl.Find(nil).Sort("-count").Limit(numTags).All(&result)
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(result))
	for i, r := range result {
		tags[i] = r.Tag
	}
	return tags, nil
}

func GetBooksVisited(num int, visitedColl *mgo.Collection) ([]bson.ObjectId, error) {
	var result []struct {
		Book bson.ObjectId "_id"
	}
	err := visitedColl.Find(nil).Sort("-value").Limit(num).All(&result)
	if err != nil {
		return nil, err
	}

	books := make([]bson.ObjectId, len(result))
	for i, r := range result {
		books[i] = r.Book
	}
	return books, nil
}

func GetVisits(visitsColl *mgo.Collection) ([]Visits, error) {
	var result []Visits
	err := visitsColl.Find(nil).All(&result)
	return result, err
}

type MR struct {
	database *mgo.Database
}

func NewMR(database *mgo.Database) *MR {
	m := new(MR)
	m.database = database
	return m
}

func (m *MR) UpdateTags(booksColl *mgo.Collection) error {
	var tags []struct {
		Tag   string "_id"
		Count int    "count"
	}
	err := booksColl.Pipe([]bson.M{
		{"$project": bson.M{"subject": 1}},
		{"$unwind": "$subject"},
		{"$group": bson.M{"_id": "$subject", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": TAGS_DISPLAY},
	}).All(&tags)
	if err != nil {
		return err
	}

	tagsColl := m.database.C(TAGS_COLL)
	tagsColl.DropCollection()
	for _, tag := range tags {
		err = tagsColl.Insert(tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MR) UpdateMostVisited(statsColl *mgo.Collection) error {
	return m.updateMostBooks(statsColl, "book", VISITED_COLL)
}

func (m *MR) UpdateMostDownloaded(statsColl *mgo.Collection) error {
	return m.updateMostBooks(statsColl, "download", DOWNLOADED_COLL)
}

func (m *MR) updateMostBooks(statsColl *mgo.Collection, section string, resColl string) error {
	const numDays = 30
	start := time.Now().UTC().Add(-numDays * 24 * time.Hour)

	var mr mgo.MapReduce
	mr.Map = `function() {
	              emit(this.id, 1);
	          }`
	mr.Reduce = `function(tag, vals) {
	                 var count = 0;
	                 vals.forEach(function() { count += 1; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"date": bson.M{"$gt": start}, "section": section}, statsColl, resColl)
}

func (m *MR) UpdateHourVisits(statsColl *mgo.Collection) error {
	f := func(t time.Time) time.Time {
		const span = time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 2
	spanStore := numDays * 24 * time.Hour
	return m.updateVisits(f, spanStore, HOURLY_VISITS_COLL, true)
}

func (m *MR) UpdateDayVisits(statsColl *mgo.Collection) error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 30
	spanStore := numDays * 24 * time.Hour
	return m.updateVisits(f, spanStore, DAILY_VISITS_COLL, true)
}

func (m *MR) UpdateMonthVisits(statsColl *mgo.Collection) error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.AddDate(0, 1, 1-t.Day()).Truncate(span)
	}
	const numDays = 365
	spanStore := numDays * 24 * time.Hour
	return m.updateVisits(f, spanStore, MONTHLY_VISITS_COLL, true)
}

func (m *MR) UpdateHourDownloads(statsColl *mgo.Collection) error {
	f := func(t time.Time) time.Time {
		const span = time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 2
	spanStore := numDays * 24 * time.Hour
	return m.updateVisits(f, spanStore, HOURLY_DOWNLOADS_COLL, false)
}

func (m *MR) UpdateDayDownloads(statsColl *mgo.Collection) error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 30
	spanStore := numDays * 24 * time.Hour
	return m.updateVisits(f, spanStore, DAILY_DOWNLOADS_COLL, false)
}

func (m *MR) UpdateMonthDownloads(statsColl *mgo.Collection) error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.AddDate(0, 1, 1-t.Day()).Truncate(span)
	}
	const numDays = 365
	spanStore := numDays * 24 * time.Hour
	return m.updateVisits(f, spanStore, MONTHLY_DOWNLOADS_COLL, false)
}

func (m *MR) updateVisits(incTime func(time.Time) time.Time, spanStore time.Duration, coll string, useSession bool) error {
	storeColl := m.database.C(coll)
	start := m.calculateStart(spanStore, storeColl)
	for start.Before(time.Now().UTC()) {
		stop := incTime(start)

		var count int
		var err error
		if useSession {
			count = m.countVisits(start, stop)
		} else {
			count, err = m.countDownloads(start, stop)
		}
		if err != nil {
			return err
		}

		err = storeColl.Insert(bson.M{"date": start, "count": count})
		if err != nil {
			return err
		}

		start = stop
	}

	_, err := storeColl.RemoveAll(bson.M{"date": bson.M{"$lt": time.Now().UTC().Add(-spanStore)}})
	return err
}

func (m *MR) calculateStart(spanStore time.Duration, storeColl *mgo.Collection) time.Time {
	var date struct {
		Id   bson.ObjectId `bson:"_id"`
		Date time.Time     `bson:"date"`
	}
	err := storeColl.Find(bson.M{}).Sort("-date").One(&date)
	if err == nil {
		storeColl.RemoveId(date.Id)
		return date.Date
	}
	return time.Now().UTC().Add(-spanStore).Truncate(time.Hour)
}

func (m *MR) countVisits(start time.Time, stop time.Time) int {
	statsColl := m.database.C(STATS_COLL)
	var result struct {
		Count int "count"
	}
	err := statsColl.Pipe([]bson.M{
		{"$match": bson.M{"date": bson.M{"$gte": start, "$lt": stop}}},
		{"$group": bson.M{"_id": "$session"}},
		{"$group": bson.M{"_id": 1, "count": bson.M{"$sum": 1}}},
	}).One(&result)
	if err != nil {
		return 0
	}

	return result.Count
}

func (m *MR) countDownloads(start time.Time, stop time.Time) (int, error) {
	query := bson.M{"date": bson.M{"$gte": start, "$lt": stop}, "section": "download"}
	statsColl := m.database.C(STATS_COLL)
	return statsColl.Find(query).Count()
}

func (m *MR) update(mr *mgo.MapReduce, query bson.M, queryColl *mgo.Collection, storeColl string) error {
	metaColl := m.database.C(META_COLL)
	_, err := metaColl.RemoveAll(bson.M{"type": storeColl})
	if err != nil {
		return err
	}

	mr.Out = bson.M{"replace": storeColl}
	_, err = queryColl.Find(query).MapReduce(mr, nil)
	if err != nil {
		return err
	}

	return metaColl.Insert(bson.M{"type": storeColl})
}
