package database

import (
	log "github.com/cihub/seelog"

	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	books_coll = "books"
)

type Book struct {
	Id                  string
	Title               string
	Author              []string
	Contributor         string
	Publisher           string
	Description         string
	Subject             []string
	Date                string
	Lang                []string
	Isbn                string
	Type                string
	Format              string
	Source              string
	Relation            string
	Coverage            string
	Rights              string
	Meta                string
	FileSize            int
	Cover               bool
	Active              bool
	BadQuality          int      `bad_quality`
	BadQualityReporters []string `bad_quality_reporters`
}

func indexBooks(coll *mgo.Collection) {
	indexes := []mgo.Index{
		{
			Key:        []string{"id"},
			Unique:     true,
			Background: true,
		},
		{
			Key:        []string{"active", "-_id"},
			Background: true,
		},
		{
			Key:        []string{"active", "-bad_quality", "-_id"},
			Background: true,
		},
		// TODO: there is no weights in mgo
	}
	for _, k := range []string{"lang", "title", "author", "subject"} {
		idx := mgo.Index{
			Key:        []string{"active", k, "-_id"},
			Background: true,
		}
		indexes = append(indexes, idx)
	}

	for _, idx := range indexes {
		err := coll.EnsureIndex(idx)
		if err != nil {
			log.Error("Error indexing books: ", err)
		}
	}
}

func addBook(coll *mgo.Collection, book map[string]interface{}) error {
	book["_lang"] = metadataLang(book)
	return coll.Insert(book)
}

func getBooks(coll *mgo.Collection, query string, length int, start int) (books []Book, num int, err error) {
	return _getBooks(coll, buildQuery(query), length, start)
}

func getNewBooks(coll *mgo.Collection, length int, start int) (books []Book, num int, err error) {
	return _getBooks(coll, bson.M{"$nor": []bson.M{{"active": true}}}, length, start)
}

func _getBooks(coll *mgo.Collection, query bson.M, length int, start int) (books []Book, num int, err error) {
	sort := []string{"$textScore:score"}
	if _, present := query["bad_quality"]; present {
		sort = append(sort, "-bad_quality")
	}
	sort = append(sort, "-_id")

	q := coll.Find(query).Select(bson.M{"score": bson.M{"$meta": "textScore"}}).Sort(sort...)
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

	if lang := metadataLang(book); lang != "" {
		data["_lang"] = lang
	}
	return coll.Update(bson.M{"id": id}, bson.M{"$set": data})
}

func flagBadQuality(coll *mgo.Collection, id string, user string) error {
	b, err := getBookId(coll, id)
	if err != nil {
		return err
	}

	for _, reporter := range b.BadQualityReporters {
		if reporter == user {
			return nil
		}
	}
	return coll.Update(
		bson.M{"id": id},
		bson.M{
			"$inc":      bson.M{"bad_quality": 1},
			"$addToSet": bson.M{"bad_quality_reporters": user},
		},
	)
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
	text := ""
	query := bson.M{"active": true}
	words := strings.Split(q, " ")
	for _, w := range words {
		tag := strings.SplitN(w, ":", 2)
		if len(tag) > 1 {
			if tag[0] == "flag" {
				query[tag[1]] = bson.M{"$gt": 0}
			} else {
				query[tag[0]] = bson.RegEx{tag[1], "i"} //FIXME: this should be a list
			}
		} else {
			if len(text) != 0 {
				text += " "
			}
			text += w
		}
	}
	if len(text) > 0 {
		query["$text"] = bson.M{"$search": text}
	}
	return query
}

func metadataLang(book map[string]interface{}) string {
	lang, ok := book["lang"].([]string)
	if !ok || len(lang) == 0 || len(lang[0]) < 2 {
		return ""
	}
	return strings.ToLower(lang[0][0:2])
}
