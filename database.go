package main

import (
	"crypto/md5"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sort"
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
	Active      bool
	Keywords    []string
}

type DB struct {
	session *mgo.Session
	books   *mgo.Collection
	user    *mgo.Collection
}

func initDB() *DB {
	var err error
	d := new(DB)
	d.session, err = mgo.Dial(DB_IP)
	if err != nil {
		panic(err)
	}

	d.books = d.session.DB(DB_NAME).C(BOOKS_COLL)
	d.user = d.session.DB(DB_NAME).C(USERS_COLL)
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

func (d *DB) IncDownload(path string) error {
	return d.books.Update(bson.M{"path": path}, bson.M{"$inc": bson.M{"DownloadCount": 1}})
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

type tagsList []struct {
	Subject string "_id"
	Count   int    "value"
}

func (t tagsList) Len() int {
	return len(t)
}
func (t tagsList) Less(i, j int) bool {
	return t[i].Count > t[j].Count
}
func (t tagsList) Swap(i, j int) {
	aux := t[i]
	t[i] = t[j]
	t[j] = aux
}

func (d *DB) GetTags() (tagsList, error) {
	// TODO: cache the tags
	var mr mgo.MapReduce
	mr.Map = "function() { " +
		"if (this.active) { this.subject.forEach(function(s) { emit(s, 1); }); }" +
		"}"
	mr.Reduce = "function(tag, vals) { " +
		"var count = 0;" +
		"vals.forEach(function() { count += 1; });" +
		"return count;" +
		"}"
	var result tagsList
	_, err := d.books.Find(bson.M{"active": true}).MapReduce(&mr, &result)
	if err == nil {
		sort.Sort(result)
	}
	return result, err
}
