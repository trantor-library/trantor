package database

import (
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
	Id          string
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
	FileSize    int
	Cover       bool
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
	return
}

func getBookId(coll *mgo.Collection, id string) (Book, error) {
	var book Book
	err := coll.Find(bson.M{"id": id}).One(&book)
	return book, err
}

func deleteBook(coll *mgo.Collection, id string) error {
	return coll.Remove(bson.M{"id": id})
}

func updateBook(coll *mgo.Collection, id string, data map[string]interface{}) error {
	var book map[string]interface{}
	err := coll.Find(bson.M{"id": id}).One(&book)
	if err != nil {
		return err
	}
	for k, v := range data {
		book[k] = v
	}

	data["keywords"] = keywords(book)
	return coll.Update(bson.M{"id": id}, bson.M{"$set": data})
}

func activeBook(coll *mgo.Collection, id string) error {
	data := map[string]interface{}{"active": true}
	return coll.Update(bson.M{"id": id}, bson.M{"$set": data})
}

func isBookActive(coll *mgo.Collection, id string) bool {
	var book Book
	err := coll.Find(bson.M{"id": id}).One(&book)
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

	k = append(k, listKeywords(b["author"])...)

	publisher, _ := b["publisher"].(string)
	k = append(k, tokens(publisher)...)

	k = append(k, listKeywords(b["subject"])...)
	return
}

func listKeywords(v interface{}) (k []string) {
	list, ok := v.([]string)
	if !ok {
		list, _ := v.([]interface{})
		for _, e := range list {
			str := e.(string)
			k = append(k, tokens(str)...)
		}
		return
	}

	for _, e := range list {
		k = append(k, tokens(e)...)
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
