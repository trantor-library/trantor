package database

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

const (
	news_coll = "news"
)

type News struct {
	Date time.Time
	Text string
}

func addNews(coll *mgo.Collection, text string) error {
	var news News
	news.Text = text
	news.Date = time.Now()
	return coll.Insert(news)
}

func getNews(coll *mgo.Collection, num int, days int) (news []News, err error) {
	query := bson.M{}
	if days != 0 {
		duration := time.Duration(-24*days) * time.Hour
		date := time.Now().Add(duration)
		query = bson.M{"date": bson.M{"$gt": date}}
	}
	q := coll.Find(query).Sort("-date").Limit(num)
	err = q.All(&news)
	return
}
