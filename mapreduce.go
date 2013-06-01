package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type MR struct {
	meta    *mgo.Collection
	tags    *mgo.Collection
	hourly  *mgo.Collection
	daily   *mgo.Collection
	monthly *mgo.Collection
}

func NewMR(database *mgo.Database) *MR {
	m := new(MR)
	m.meta = database.C(META_COLL)
	m.tags = database.C(TAGS_COLL)
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

func (m *MR) GetHourVisits(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(HOURLY_VISITS_COLL, MINUTES_UPDATE_HOURLY) {
		var mr mgo.MapReduce
		mr.Map = `function() {
		              var day = Date.UTC(this.date.getUTCFullYear(),
		                                 this.date.getUTCMonth(),
		                                 this.date.getUTCDate(),
	                                         this.date.getUTCHours());
		              emit(day, 1);
		          }`
		mr.Reduce = `function(date, vals) {
				 var count = 0;
				 vals.forEach(function(v) { count += v; });
				 return count;
			     }`
		err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, HOURLY_VISITS_COLL)
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
		var mr mgo.MapReduce
		mr.Map = `function() {
		              var day = Date.UTC(this.date.getUTCFullYear(),
		                                 this.date.getUTCMonth(),
		                                 this.date.getUTCDate());
		              emit(day, 1);
		          }`
		mr.Reduce = `function(date, vals) {
				 var count = 0;
				 vals.forEach(function(v) { count += v; });
				 return count;
			     }`
		err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, DAILY_VISITS_COLL)
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
		var mr mgo.MapReduce
		mr.Map = `function() {
		              var day = Date.UTC(this.date.getUTCFullYear(),
		                                 this.date.getUTCMonth());
		              emit(day, 1);
		          }`
		mr.Reduce = `function(date, vals) {
				 var count = 0;
				 vals.forEach(function(v) { count += v; });
				 return count;
			     }`
		err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, MONTHLY_VISITS_COLL)
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
