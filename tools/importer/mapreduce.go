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
	err := tagsColl.Find(nil).Sort("-value").Limit(numTags).All(&result)
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
	var mr mgo.MapReduce
	mr.Map = `function() {
	              if (this.subject) {
	                  this.subject.forEach(function(s) { emit(s, 1); });
	              }
	          }`
	mr.Reduce = `function(tag, vals) {
	                 var count = 0;
	                 vals.forEach(function() { count += 1; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"active": true}, booksColl, TAGS_COLL)
}

func (m *MR) UpdateMostVisited(statsColl *mgo.Collection) error {
	var mr mgo.MapReduce
	mr.Map = `function() {
	              if (this.id) {
	                  emit(this.id, 1);
	              }
	          }`
	mr.Reduce = `function(tag, vals) {
	                 var count = 0;
	                 vals.forEach(function() { count += 1; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"section": "book"}, statsColl, VISITED_COLL)
}

func (m *MR) UpdateMostDownloaded(statsColl *mgo.Collection) error {
	var mr mgo.MapReduce
	mr.Map = `function() {
	              emit(this.id, 1);
	          }`
	mr.Reduce = `function(tag, vals) {
	                 var count = 0;
	                 vals.forEach(function() { count += 1; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"section": "download"}, statsColl, DOWNLOADED_COLL)
}

func (m *MR) UpdateHourVisits(statsColl *mgo.Collection) error {
	const numDays = 2
	start := time.Now().UTC().Add(-numDays * 24 * time.Hour)

	const reduce = `function(date, vals) {
	                    var count = 0;
	                    vals.forEach(function(v) { count += v; });
	                    return count;
	                }`
	var mr mgo.MapReduce
	mr.Map = `function() {
	             var date = Date.UTC(this.date.getUTCFullYear(),
	                                 this.date.getUTCMonth(),
	                                 this.date.getUTCDate(),
	                                 this.date.getUTCHours());
	             emit({date: date, session: this.session}, 1);
	          }`
	mr.Reduce = reduce
	err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, HOURLY_VISITS_COLL+"_raw")
	if err != nil {
		return err
	}
	var mr2 mgo.MapReduce
	mr2.Map = `function() {
	               emit(this['_id']['date'], 1);
	           }`
	mr2.Reduce = reduce
	hourly_raw := m.database.C(HOURLY_VISITS_COLL + "_raw")
	return m.update(&mr2, bson.M{}, hourly_raw, HOURLY_VISITS_COLL)
}

func (m *MR) UpdateDayVisits(statsColl *mgo.Collection) error {
	const numDays = 30
	start := time.Now().UTC().Add(-numDays * 24 * time.Hour).Truncate(24 * time.Hour)

	const reduce = `function(date, vals) {
	                    var count = 0;
	                    vals.forEach(function(v) { count += v; });
	                    return count;
	                }`
	var mr mgo.MapReduce
	mr.Map = `function() {
	              var date = Date.UTC(this.date.getUTCFullYear(),
	                                  this.date.getUTCMonth(),
	                                  this.date.getUTCDate());
	              emit({date: date, session: this.session}, 1);
	          }`
	mr.Reduce = reduce
	err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, DAILY_VISITS_COLL+"_raw")
	if err != nil {
		return err
	}
	var mr2 mgo.MapReduce
	mr2.Map = `function() {
	              emit(this['_id']['date'], 1);
	           }`
	mr2.Reduce = reduce
	daily_raw := m.database.C(DAILY_VISITS_COLL + "_raw")
	return m.update(&mr2, bson.M{}, daily_raw, DAILY_VISITS_COLL)
}

func (m *MR) UpdateMonthVisits(statsColl *mgo.Collection) error {
	const numDays = 365

	start := time.Now().UTC().Add(-numDays * 24 * time.Hour).Truncate(24 * time.Hour)

	const reduce = `function(date, vals) {
	                    var count = 0;
	                    vals.forEach(function(v) { count += v; });
	                    return count;
	                }`
	var mr mgo.MapReduce
	mr.Map = `function() {
	              var date = Date.UTC(this.date.getUTCFullYear(),
	                                  this.date.getUTCMonth());
	              emit({date: date, session: this.session}, 1);
	          }`
	mr.Reduce = reduce
	err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, MONTHLY_VISITS_COLL+"_raw")
	if err != nil {
		return err
	}
	var mr2 mgo.MapReduce
	mr2.Map = `function() {
	               emit(this['_id']['date'], 1);
	           }`
	mr2.Reduce = reduce
	monthly_raw := m.database.C(MONTHLY_VISITS_COLL + "_raw")
	return m.update(&mr2, bson.M{}, monthly_raw, MONTHLY_VISITS_COLL)
}

func (m *MR) UpdateHourDownloads(statsColl *mgo.Collection) error {
	const numDays = 2
	start := time.Now().UTC().Add(-numDays * 24 * time.Hour)

	var mr mgo.MapReduce
	mr.Map = `function() {
	              if (this.section == "download") {
	                  var date = Date.UTC(this.date.getUTCFullYear(),
	                                      this.date.getUTCMonth(),
	                                      this.date.getUTCDate(),
	                                      this.date.getUTCHours());
	                  emit(date, 1);
	              }
                  }`
	mr.Reduce = `function(date, vals) {
	                 var count = 0;
	                 vals.forEach(function(v) { count += v; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, HOURLY_DOWNLOADS_COLL)
}

func (m *MR) UpdateDayDownloads(statsColl *mgo.Collection) error {
	const numDays = 30
	start := time.Now().UTC().Add(-numDays * 24 * time.Hour).Truncate(24 * time.Hour)

	var mr mgo.MapReduce
	mr.Map = `function() {
	              if (this.section == "download") {
	                  var date = Date.UTC(this.date.getUTCFullYear(),
	                                      this.date.getUTCMonth(),
	                                      this.date.getUTCDate());
	                  emit(date, 1);
	              }
	          }`
	mr.Reduce = `function(date, vals) {
	                 var count = 0;
	                 vals.forEach(function(v) { count += v; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, DAILY_DOWNLOADS_COLL)
}

func (m *MR) UpdateMonthDownloads(statsColl *mgo.Collection) error {
	const numDays = 365

	start := time.Now().UTC().Add(-numDays * 24 * time.Hour).Truncate(24 * time.Hour)

	var mr mgo.MapReduce
	mr.Map = `function() {
	              if (this.section == "download") {
	                  var date = Date.UTC(this.date.getUTCFullYear(),
	                                      this.date.getUTCMonth());
	                  emit(date, 1);
	              }
	          }`
	mr.Reduce = `function(date, vals) {
	                 var count = 0;
	                 vals.forEach(function(v) { count += v; });
	                 return count;
	             }`
	return m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, MONTHLY_DOWNLOADS_COLL)
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

func (m *MR) isOutdated(coll string, minutes float64) bool {
	var result struct {
		Id bson.ObjectId `bson:"_id"`
	}
	metaColl := m.database.C(META_COLL)
	err := metaColl.Find(bson.M{"type": coll}).One(&result)
	if err != nil {
		return true
	}

	lastUpdate := result.Id.Time()
	return time.Since(lastUpdate).Minutes() > minutes
}
