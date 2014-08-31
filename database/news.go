package database

import (
	log "github.com/cihub/seelog"

	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	news_coll = "news"
)

type News struct {
	Date time.Time
	Text string
}

func indexNews(coll *mgo.Collection) {
	idx := mgo.Index{
		Key:        []string{"-date"},
		Unique:     true,
		Background: true,
	}
	err := coll.EnsureIndex(idx)
	if err != nil {
		log.Error("Error indexing news: ", err)
	}
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
