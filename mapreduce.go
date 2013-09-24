package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type MR struct {
	database *mgo.Database
}

func NewMR(database *mgo.Database) *MR {
	m := new(MR)
	m.database = database
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
	tagsColl := m.database.C(TAGS_COLL)
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
	visitedColl := m.database.C(VISITED_COLL)
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
	downloadedColl := m.database.C(DOWNLOADED_COLL)
	err := downloadedColl.Find(nil).Sort("-value").Limit(num).All(&result)
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
		hourly_raw := m.database.C(HOURLY_VISITS_COLL + "_raw")
		err = m.update(&mr2, bson.M{}, hourly_raw, HOURLY_VISITS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	hourlyColl := m.database.C(HOURLY_VISITS_COLL)
	err := hourlyColl.Find(nil).All(&result)
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
		daily_raw := m.database.C(DAILY_VISITS_COLL + "_raw")
		err = m.update(&mr2, bson.M{}, daily_raw, DAILY_VISITS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	dailyColl := m.database.C(DAILY_VISITS_COLL)
	err := dailyColl.Find(nil).All(&result)
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
		monthly_raw := m.database.C(MONTHLY_VISITS_COLL + "_raw")
		err = m.update(&mr2, bson.M{}, monthly_raw, MONTHLY_VISITS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	monthlyColl := m.database.C(MONTHLY_VISITS_COLL)
	err := monthlyColl.Find(nil).All(&result)
	return result, err
}

func (m *MR) GetHourDownloads(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(HOURLY_DOWNLOADS_COLL, MINUTES_UPDATE_HOURLY) {
		const reduce = `function(date, vals) {
		                    var count = 0;
		                    vals.forEach(function(v) { count += v; });
		                    return count;
		                }`
		var mr mgo.MapReduce
		mr.Map = `function() {
		              if (this.section == "download") {
						  var date = Date.UTC(this.date.getUTCFullYear(),
											  this.date.getUTCMonth(),
											  this.date.getUTCDate(),
											  this.date.getUTCHours());
						  emit({date: date}, 1);
					  }
		          }`
		mr.Reduce = reduce
		err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, HOURLY_DOWNLOADS_COLL+"_raw")
		if err != nil {
			return nil, err
		}
		var mr2 mgo.MapReduce
		mr2.Map = `function() {
		               emit(this['_id']['date'], 1);
		           }`
		mr2.Reduce = reduce
		hourly_raw := m.database.C(HOURLY_DOWNLOADS_COLL + "_raw")
		err = m.update(&mr2, bson.M{}, hourly_raw, HOURLY_DOWNLOADS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	hourlyColl := m.database.C(HOURLY_DOWNLOADS_COLL)
	err := hourlyColl.Find(nil).All(&result)
	return result, err
}

func (m *MR) GetDayDowloads(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(DAILY_DOWNLOADS_COLL, MINUTES_UPDATE_DAILY) {
		const reduce = `function(date, vals) {
		                    var count = 0;
		                    vals.forEach(function(v) { count += v; });
		                    return count;
		                }`
		var mr mgo.MapReduce
		mr.Map = `function() {
		              if (this.section == "download") {
						  var date = Date.UTC(this.date.getUTCFullYear(),
											  this.date.getUTCMonth(),
											  this.date.getUTCDate());
						  emit({date: date, session: this.session}, 1);
					  }
		          }`
		mr.Reduce = reduce
		err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, DAILY_DOWNLOADS_COLL+"_raw")
		if err != nil {
			return nil, err
		}
		var mr2 mgo.MapReduce
		mr2.Map = `function() {
		               emit(this['_id']['date'], 1);
		           }`
		mr2.Reduce = reduce
		daily_raw := m.database.C(DAILY_DOWNLOADS_COLL + "_raw")
		err = m.update(&mr2, bson.M{}, daily_raw, DAILY_DOWNLOADS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	dailyColl := m.database.C(DAILY_DOWNLOADS_COLL)
	err := dailyColl.Find(nil).All(&result)
	return result, err
}

func (m *MR) GetMonthDowloads(start time.Time, statsColl *mgo.Collection) ([]Visits, error) {
	if m.isOutdated(MONTHLY_DOWNLOADS_COLL, MINUTES_UPDATE_MONTHLY) {
		const reduce = `function(date, vals) {
		                    var count = 0;
		                    vals.forEach(function(v) { count += v; });
		                    return count;
		                }`
		var mr mgo.MapReduce
		mr.Map = `function() {
		              if (this.section == "download") {
						  var date = Date.UTC(this.date.getUTCFullYear(),
											  this.date.getUTCMonth());
						  emit({date: date, session: this.session}, 1);
			          }
		          }`
		mr.Reduce = reduce
		err := m.update(&mr, bson.M{"date": bson.M{"$gte": start}}, statsColl, MONTHLY_DOWNLOADS_COLL+"_raw")
		if err != nil {
			return nil, err
		}
		var mr2 mgo.MapReduce
		mr2.Map = `function() {
		               emit(this['_id']['date'], 1);
		           }`
		mr2.Reduce = reduce
		monthly_raw := m.database.C(MONTHLY_DOWNLOADS_COLL + "_raw")
		err = m.update(&mr2, bson.M{}, monthly_raw, MONTHLY_DOWNLOADS_COLL)
		if err != nil {
			return nil, err
		}
	}

	var result []Visits
	monthlyColl := m.database.C(MONTHLY_DOWNLOADS_COLL)
	err := monthlyColl.Find(nil).All(&result)
	return result, err
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
