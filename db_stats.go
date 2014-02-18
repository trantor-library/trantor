package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type DBUpdate struct {
	src *mgo.Collection
	dst *mgo.Collection
}

func GetTags(tagsColl *mgo.Collection) ([]string, error) {
	var result []struct {
		Tag string "_id"
	}
	err := tagsColl.Find(nil).Sort("-count").All(&result)
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(result))
	for i, r := range result {
		tags[i] = r.Tag
	}
	return tags, nil
}

func GetBooksVisited(visitedColl *mgo.Collection) ([]bson.ObjectId, error) {
	var result []struct {
		Book bson.ObjectId "_id"
	}
	err := visitedColl.Find(nil).Sort("-count").All(&result)
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

func (u *DBUpdate) UpdateTags() error {
	var tags []struct {
		Tag   string "_id"
		Count int    "count"
	}
	err := u.src.Pipe([]bson.M{
		{"$project": bson.M{"subject": 1}},
		{"$unwind": "$subject"},
		{"$group": bson.M{"_id": "$subject", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": TAGS_DISPLAY},
	}).All(&tags)
	if err != nil {
		return err
	}

	u.dst.DropCollection()
	for _, tag := range tags {
		err = u.dst.Insert(tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *DBUpdate) UpdateMostBooks(section string) error {
	const numDays = 30
	start := time.Now().UTC().Add(-numDays * 24 * time.Hour)

	var books []struct {
		Book  string "_id"
		Count int    "count"
	}
	err := u.src.Pipe([]bson.M{
		{"$match": bson.M{"date": bson.M{"$gt": start}, "section": section}},
		{"$project": bson.M{"id": 1}},
		{"$group": bson.M{"_id": "$id", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
		{"$limit": BOOKS_FRONT_PAGE},
	}).All(&books)
	if err != nil {
		return err
	}

	u.dst.DropCollection()
	for _, book := range books {
		err = u.dst.Insert(book)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *DBUpdate) UpdateHourVisits() error {
	f := func(t time.Time) time.Time {
		const span = time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 2
	spanStore := numDays * 24 * time.Hour
	return u.updateVisits(f, spanStore, HOURLY_VISITS_COLL, true)
}

func (u *DBUpdate) UpdateDayVisits() error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 30
	spanStore := numDays * 24 * time.Hour
	return u.updateVisits(f, spanStore, DAILY_VISITS_COLL, true)
}

func (u *DBUpdate) UpdateMonthVisits() error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.AddDate(0, 1, 1-t.Day()).Truncate(span)
	}
	const numDays = 365
	spanStore := numDays * 24 * time.Hour
	return u.updateVisits(f, spanStore, MONTHLY_VISITS_COLL, true)
}

func (u *DBUpdate) UpdateHourDownloads() error {
	f := func(t time.Time) time.Time {
		const span = time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 2
	spanStore := numDays * 24 * time.Hour
	return u.updateVisits(f, spanStore, HOURLY_DOWNLOADS_COLL, false)
}

func (u *DBUpdate) UpdateDayDownloads() error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.Add(span).Truncate(span)
	}
	const numDays = 30
	spanStore := numDays * 24 * time.Hour
	return u.updateVisits(f, spanStore, DAILY_DOWNLOADS_COLL, false)
}

func (u *DBUpdate) UpdateMonthDownloads() error {
	f := func(t time.Time) time.Time {
		const span = 24 * time.Hour
		return t.AddDate(0, 1, 1-t.Day()).Truncate(span)
	}
	const numDays = 365
	spanStore := numDays * 24 * time.Hour
	return u.updateVisits(f, spanStore, MONTHLY_DOWNLOADS_COLL, false)
}

func (u *DBUpdate) updateVisits(incTime func(time.Time) time.Time, spanStore time.Duration, coll string, useSession bool) error {
	start := u.calculateStart(spanStore)
	for start.Before(time.Now().UTC()) {
		stop := incTime(start)

		var count int
		var err error
		if useSession {
			count = u.countVisits(start, stop)
		} else {
			count, err = u.countDownloads(start, stop)
		}
		if err != nil {
			return err
		}

		err = u.dst.Insert(bson.M{"date": start, "count": count})
		if err != nil {
			return err
		}

		start = stop
	}

	_, err := u.dst.RemoveAll(bson.M{"date": bson.M{"$lt": time.Now().UTC().Add(-spanStore)}})
	return err
}

func (u *DBUpdate) calculateStart(spanStore time.Duration) time.Time {
	var date struct {
		Id   bson.ObjectId `bson:"_id"`
		Date time.Time     `bson:"date"`
	}
	err := u.dst.Find(bson.M{}).Sort("-date").One(&date)
	if err == nil {
		u.dst.RemoveId(date.Id)
		return date.Date
	}
	return time.Now().UTC().Add(-spanStore).Truncate(time.Hour)
}

func (u *DBUpdate) countVisits(start time.Time, stop time.Time) int {
	var result struct {
		Count int "count"
	}
	err := u.src.Pipe([]bson.M{
		{"$match": bson.M{"date": bson.M{"$gte": start, "$lt": stop}}},
		{"$group": bson.M{"_id": "$session"}},
		{"$group": bson.M{"_id": 1, "count": bson.M{"$sum": 1}}},
	}).One(&result)
	if err != nil {
		return 0
	}

	return result.Count
}

func (u *DBUpdate) countDownloads(start time.Time, stop time.Time) (int, error) {
	query := bson.M{"date": bson.M{"$gte": start, "$lt": stop}, "section": "download"}
	return u.src.Find(query).Count()
}
