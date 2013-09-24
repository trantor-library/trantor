package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type MR struct {
	meta        *mgo.Collection
	tags        *mgo.Collection
	visited     *mgo.Collection
	downloaded  *mgo.Collection
	hourly_raw  *mgo.Collection
	daily_raw   *mgo.Collection
	monthly_raw *mgo.Collection
	hourly      *mgo.Collection
	daily       *mgo.Collection
	monthly     *mgo.Collection
}

func NewMR(database *mgo.Database) *MR {
	m := new(MR)
	m.meta = database.C(META_COLL)
	m.tags = database.C(TAGS_COLL)
	m.visited = database.C(VISITED_COLL)
	m.downloaded = database.C(DOWNLOADED_COLL)
	m.hourly_raw = database.C(HOURLY_VISITS_COLL + "_raw")
	m.daily_raw = database.C(DAILY_VISITS_COLL + "_raw")
	m.monthly_raw = database.C(MONTHLY_VISITS_COLL + "_raw")
	m.hourly = database.C(HOURLY_VISITS_COLL)
	m.daily = database.C(DAILY_VISITS_COLL)
	m.monthly = database.C(MONTHLY_VISITS_COLL)
	return m
}

func (m *MR) GetTags(numTags int, booksColl *mgo.Collection) ([]string, error) {
	if m.isOutdated(TAGS_COLL, MINUTES_UPDATE_TAGS) {
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
		err := m.update(&mr, bson.M{"active": true}, booksColl, TAGS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []struct {
		Tag string "_id"
	}
	err := m.tags.Find(nil).Sort("-value").Limit(numTags).All(&result)
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(result))
	for i, r := range result {
		tags[i] = r.Tag
	}
	return tags, nil
}

func (m *MR) GetMostVisited(num int, statsColl *mgo.Collection) ([]bson.ObjectId, error) {
	if m.isOutdated(VISITED_COLL, MINUTES_UPDATE_VISITED) {
		var mr mgo.MapReduce
		mr.Map = `function() {
		              emit(this.id, 1);
			  }`
		mr.Reduce = `function(tag, vals) {
				 var count = 0;
				 vals.forEach(function() { count += 1; });
				 return count;
			     }`
		err := m.update(&mr, bson.M{"section": "book"}, statsColl, VISITED_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []struct {
		Book bson.ObjectId "_id"
	}
	err := m.visited.Find(nil).Sort("-value").Limit(num).All(&result)
	if err != nil {
		return nil, err
	}

	books := make([]bson.ObjectId, len(result))
	for i, r := range result {
		books[i] = r.Book
	}
	return books, nil
}

func (m *MR) GetMostDownloaded(num int, statsColl *mgo.Collection) ([]bson.ObjectId, error) {
	if m.isOutdated(DOWNLOADED_COLL, MINUTES_UPDATE_DOWNLOADED) {
		var mr mgo.MapReduce
		mr.Map = `function() {
		              emit(this.id, 1);
			  }`
		mr.Reduce = `function(tag, vals) {
				 var count = 0;
				 vals.forEach(function() { count += 1; });
				 return count;
			     }`
		err := m.update(&mr, bson.M{"section": "download"}, statsColl, DOWNLOADED_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []struct {
		Book bson.ObjectId "_id"
	}
	err := m.downloaded.Find(nil).Sort("-value").Limit(num).All(&result)
	if err != nil {
		return nil, err
	}

	books := make([]bson.ObjectId, len(result))
	for i, r := range result {
		books[i] = r.Book
	}
	return books, nil
}

func (m *MR) GetHourVisits(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(HOURLY_VISITS_COLL, MINUTES_UPDATE_HOURLY) {
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
			return nil, err
		}
		var mr2 mgo.MapReduce
		mr2.Map = `function() {
		               emit(this['_id']['date'], 1);
		           }`
		mr2.Reduce = reduce
		err = m.update(&mr2, bson.M{}, m.hourly_raw, HOURLY_VISITS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	err := m.hourly.Find(nil).All(&result)
	return result, err
}

func (m *MR) GetDayVisits(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(DAILY_VISITS_COLL, MINUTES_UPDATE_DAILY) {
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
			return nil, err
		}
		var mr2 mgo.MapReduce
		mr2.Map = `function() {
		               emit(this['_id']['date'], 1);
		           }`
		mr2.Reduce = reduce
		err = m.update(&mr2, bson.M{}, m.daily_raw, DAILY_VISITS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	err := m.daily.Find(nil).All(&result)
	return result, err
}

func (m *MR) GetMonthVisits(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(MONTHLY_VISITS_COLL, MINUTES_UPDATE_MONTHLY) {
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
			return nil, err
		}
		var mr2 mgo.MapReduce
		mr2.Map = `function() {
		               emit(this['_id']['date'], 1);
		           }`
		mr2.Reduce = reduce
		err = m.update(&mr2, bson.M{}, m.monthly_raw, MONTHLY_VISITS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	err := m.monthly.Find(nil).All(&result)
	return result, err
}

func (m *MR) update(mr *mgo.MapReduce, query bson.M, queryColl *mgo.Collection, storeColl string) error {
	_, err := m.meta.RemoveAll(bson.M{"type": storeColl})
	if err != nil {
		return err
	}

	mr.Out = bson.M{"replace": storeColl}
	_, err = queryColl.Find(query).MapReduce(mr, nil)
	if err != nil {
		return err
	}

	return m.meta.Insert(bson.M{"type": storeColl})
}

func (m *MR) isOutdated(coll string, minutes float64) bool {
	var result struct {
		Id bson.ObjectId `bson:"_id"`
	}
	err := m.meta.Find(bson.M{"type": coll}).One(&result)
	if err != nil {
		return true
	}

	lastUpdate := result.Id.Time()
	return time.Since(lastUpdate).Minutes() > minutes
}
