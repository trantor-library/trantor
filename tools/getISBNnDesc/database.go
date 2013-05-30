package main

import (
	"crypto/md5"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

const (
	META_TYPE_TAGS = "tags updated"
)

var db *DB

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
	Cover       bson.ObjectId
	CoverSmall  bson.ObjectId
	Active      bool
	Keywords    []string
}

type DB struct {
	session *mgo.Session
	meta    *mgo.Collection
	books   *mgo.Collection
	tags    *mgo.Collection
	user    *mgo.Collection
	stats   *mgo.Collection
}

func initDB() *DB {
	var err error
	d := new(DB)
	d.session, err = mgo.Dial(DB_IP)
	if err != nil {
		panic(err)
	}

	database := d.session.DB(DB_NAME)
	d.meta = database.C(META_COLL)
	d.books = database.C(BOOKS_COLL)
	d.tags = database.C(TAGS_COLL)
	d.user = database.C(USERS_COLL)
	d.stats = database.C(STATS_COLL)
	return d
}

func (d *DB) Close() {
	d.session.Close()
}

func md5Pass(pass string) []byte {
	h := md5.New()
	hash := h.Sum(([]byte)(PASS_SALT + pass))
	return hash
}

func (d *DB) SetPassword(user string, pass string) error {
	hash := md5Pass(pass)
	return d.user.Update(bson.M{"user": user}, bson.M{"$set": bson.M{"pass": hash}})
}

func (d *DB) UserValid(user string, pass string) bool {
	hash := md5Pass(pass)
	n, err := d.user.Find(bson.M{"user": user, "pass": hash}).Count()
	if err != nil {
		return false
	}
	return n != 0
}

func (d *DB) InsertStats(stats interface{}) error {
	return d.stats.Insert(stats)
}

func (d *DB) InsertBook(book interface{}) error {
	return d.books.Insert(book)
}

func (d *DB) RemoveBook(id bson.ObjectId) error {
	return d.books.Remove(bson.M{"_id": id})
}

func (d *DB) UpdateBook(id bson.ObjectId, data interface{}) error {
	return d.books.Update(bson.M{"_id": id}, bson.M{"$set": data})
}

func (d *DB) IncVisit(id bson.ObjectId) error {
	return d.books.Update(bson.M{"_id": id}, bson.M{"$inc": bson.M{"VisitsCount": 1}})
}

func (d *DB) IncDownload(id bson.ObjectId) error {
	return d.books.Update(bson.M{"_id": id}, bson.M{"$inc": bson.M{"DownloadCount": 1}})
}

/* optional parameters: length and start index
 *
 * Returns: list of books, number found and err
 */
func (d *DB) GetBooks(query bson.M, r ...int) (books []Book, num int, err error) {
	var start, length int
	if len(r) > 0 {
		length = r[0]
		if len(r) > 1 {
			start = r[1]
		}
	}
	q := d.books.Find(query).Sort("-_id")
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

/* Get the most visited books
 */
func (d *DB) GetVisitedBooks(num int) (books []Book, err error) {
	var q *mgo.Query
	q = d.books.Find(bson.M{"active": true}).Sort("-VisitsCount").Limit(num)
	err = q.All(&books)
	for i, b := range books {
		books[i].Id = bson.ObjectId(b.Id).Hex()
	}
	return
}

/* Get the most downloaded books
 */
func (d *DB) GetDownloadedBooks(num int) (books []Book, err error) {
	var q *mgo.Query
	q = d.books.Find(bson.M{"active": true}).Sort("-DownloadCount").Limit(num)
	err = q.All(&books)
	for i, b := range books {
		books[i].Id = bson.ObjectId(b.Id).Hex()
	}
	return
}

/* optional parameters: length and start index
 *
 * Returns: list of books, number found and err
 */
func (d *DB) GetNewBooks(r ...int) (books []Book, num int, err error) {
	return d.GetBooks(bson.M{"$nor": []bson.M{{"active": true}}}, r...)
}

func (d *DB) BookActive(id bson.ObjectId) bool {
	var book Book
	err := d.books.Find(bson.M{"_id": id}).One(&book)
	if err != nil {
		return false
	}
	return book.Active
}

func (d *DB) GetFS(prefix string) *mgo.GridFS {
	return d.session.DB(DB_NAME).GridFS(prefix)
}

func (d *DB) areTagsOutdated() bool {
	var result struct {
		Id bson.ObjectId `bson:"_id"`
	}
	err := d.meta.Find(bson.M{"type": META_TYPE_TAGS}).One(&result)
	if err != nil {
		return true
	}

	lastUpdate := result.Id.Time()
	return time.Since(lastUpdate).Minutes() > MINUTES_UPDATE_TAGS
}

func (d *DB) updateTags() error {
	_, err := d.meta.RemoveAll(bson.M{"type": META_TYPE_TAGS})
	if err != nil {
		return err
	}

	var mr mgo.MapReduce
	mr.Map = "function() { " +
		"if (this.active) { this.subject.forEach(function(s) { emit(s, 1); }); }" +
		"}"
	mr.Reduce = "function(tag, vals) { " +
		"var count = 0;" +
		"vals.forEach(function() { count += 1; });" +
		"return count;" +
		"}"
	mr.Out = bson.M{"replace": TAGS_COLL}
	_, err = d.books.Find(bson.M{"active": true}).MapReduce(&mr, nil)
	if err != nil {
		return err
	}

	return d.meta.Insert(bson.M{"type": META_TYPE_TAGS})
}

func (d *DB) GetTags(numTags int) ([]string, error) {
	if d.areTagsOutdated() {
		err := d.updateTags()
		if err != nil {
			return nil, err
		}
	}

	var result []struct {
		Tag string "_id"
	}
	err := d.tags.Find(nil).Sort("-value").Limit(numTags).All(&result)
	if err != nil {
		return nil, err
	}
	tags := make([]string, len(result))
	for i, r := range result {
		tags[i] = r.Tag
	}
	return tags, nil
}
