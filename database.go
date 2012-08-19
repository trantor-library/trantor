package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Book struct {
	Id          string `bson:"_id"`
	Title       string
	Author      []string
	Contributor string
	Publisher   string
	Description string
	Subject     []string
	Date        string
	Lang        []string
	Type        string
	Format      string
	Source      string
	Relation    string
	Coverage    string
	Rights      string
	Meta        string
	Path        string
	Cover       string
	CoverSmall  string
	Keywords    []string
}

func GetBook(coll *mgo.Collection, query bson.M) ([]Book, error) {
	var books []Book
	err := coll.Find(query).All(&books)
	for i, b := range books {
		books[i].Id = bson.ObjectId(b.Id).Hex()
	}
	return books, err

}
