package database

import (
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkgs.com/unidecode.v1"
	"strings"
	"unicode"
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

func addBook(coll *mgo.Collection, book map[string]interface{}) error {
	book["keywords"] = keywords(book)
	return coll.Insert(book)
}

func getBooks(coll *mgo.Collection, query string, length int, start int) (books []Book, num int, err error) {
	return _getBooks(coll, buildQuery(query), length, start)
}

func getNewBooks(coll *mgo.Collection, length int, start int) (books []Book, num int, err error) {
	return _getBooks(coll, bson.M{"$nor": []bson.M{{"active": true}}}, length, start)
}

func _getBooks(coll *mgo.Collection, query bson.M, length int, start int) (books []Book, num int, err error) {
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

func deleteBook(coll *mgo.Collection, id string) error {
	return coll.RemoveId(bson.ObjectIdHex(id))
}

func updateBook(coll *mgo.Collection, id string, data map[string]interface{}) error {
	data["keywords"] = keywords(data)
	return coll.UpdateId(bson.ObjectIdHex(id), bson.M{"$set": data})
}

func bookActive(coll *mgo.Collection, id string) bool {
	var book Book
	err := coll.FindId(bson.ObjectIdHex(id)).One(&book)
	if err != nil {
		return false
	}
	return book.Active
}

func buildQuery(q string) bson.M {
	var keywords []string
	query := bson.M{"active": true}
	words := strings.Split(q, " ")
	for _, w := range words {
		tag := strings.SplitN(w, ":", 2)
		if len(tag) > 1 {
			query[tag[0]] = bson.RegEx{tag[1], "i"}
		} else {
			toks := tokens(w)
			keywords = append(keywords, toks...)
		}
	}
	if len(keywords) > 0 {
		query["keywords"] = bson.M{"$all": keywords}
	}
	return query
}

func keywords(b map[string]interface{}) (k []string) {
	title, _ := b["title"].(string)
	k = tokens(title)
	author, _ := b["author"].([]string)
	for _, a := range author {
		k = append(k, tokens(a)...)
	}
	publisher, _ := b["publisher"].(string)
	k = append(k, tokens(publisher)...)
	subject, _ := b["subject"].([]string)
	for _, s := range subject {
		k = append(k, tokens(s)...)
	}
	return
}

func tokens(str string) []string {
	str = unidecode.Unidecode(str)
	str = strings.ToLower(str)
	f := func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsPunct(r) || unicode.IsSpace(r)
	}
	return strings.FieldsFunc(str, f)
}
