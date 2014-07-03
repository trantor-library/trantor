package database

import (
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	books_coll = "books"
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
	Isbn        string
	Type        string
	Format      string
	Source      string
	Relation    string
	Coverage    string
	Rights      string
	Meta        string
	File        bson.ObjectId
	FileSize    int
	Cover       bson.ObjectId
	CoverSmall  bson.ObjectId
	Active      bool
	Keywords    []string
}

func addBook(coll *mgo.Collection, book interface{}) error {
	return coll.Insert(book)
}

func getBooks(coll *mgo.Collection, query bson.M, length int, start int) (books []Book, num int, err error) {
	q := coll.Find(query).Sort("-_id")
	num, err = q.Count()
	if err != nil {
		return
	}
	if start != 0 {
		q = q.Skip(start)
	}
	if length != 0 {
		q = q.Limit(length)
	}

	err = q.All(&books)
	for i, b := range books {
		books[i].Id = bson.ObjectId(b.Id).Hex()
	}
	return
}

func getBookId(coll *mgo.Collection, id string) (Book, error) {
	var book Book
	if !bson.IsObjectIdHex(id) {
		return book, errors.New("Not valid book id")
	}

	err := coll.FindId(bson.ObjectIdHex(id)).One(&book)
	book.Id = bson.ObjectId(book.Id).Hex()
	return book, err
}

func deleteBook(coll *mgo.Collection, id bson.ObjectId) error {
	return coll.Remove(bson.M{"_id": id})
}

func updateBook(coll *mgo.Collection, id bson.ObjectId, data interface{}) error {
	return coll.Update(bson.M{"_id": id}, bson.M{"$set": data})
}

func bookActive(coll *mgo.Collection, id bson.ObjectId) bool {
	var book Book
	err := coll.Find(bson.M{"_id": id}).One(&book)
	if err != nil {
		return false
	}
	return book.Active
}
