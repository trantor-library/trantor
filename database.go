package main

import (
	"crypto/md5"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
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

type News struct {
	Date time.Time
	Text string
}

type DB struct {
	session *mgo.Session
}

func initDB() *DB {
	var err error
	d := new(DB)
	d.session, err = mgo.Dial(DB_IP)
	if err != nil {
		panic(err)
	}
	return d
}

func (d *DB) Close() {
	d.session.Close()
}

func (d *DB) Copy() *DB {
	dbCopy := new(DB)
	dbCopy.session = d.session.Copy()
	return dbCopy
}

func md5Pass(pass string) []byte {
	h := md5.New()
	hash := h.Sum(([]byte)(PASS_SALT + pass))
	return hash
}

func (d *DB) SetPassword(user string, pass string) error {
	hash := md5Pass(pass)
	userColl := d.session.DB(DB_NAME).C(USERS_COLL)
	return userColl.Update(bson.M{"user": user}, bson.M{"$set": bson.M{"pass": hash}})
}

func (d *DB) UserValid(user string, pass string) bool {
	hash := md5Pass(pass)
	userColl := d.session.DB(DB_NAME).C(USERS_COLL)
	n, err := userColl.Find(bson.M{"user": user, "pass": hash}).Count()
	if err != nil {
		return false
	}
	return n != 0
}

func (d *DB) UserRole(user string) string {
	type result struct {
		Role string
	}
	res := result{}
	userColl := d.session.DB(DB_NAME).C(USERS_COLL)
	err := userColl.Find(bson.M{"user": user}).One(&res)
	if err != nil {
		return ""
	}
	return res.Role
}

func (d *DB) AddNews(text string) error {
	var news News
	news.Text = text
	news.Date = time.Now()
	newsColl := d.session.DB(DB_NAME).C(NEWS_COLL)
	return newsColl.Insert(news)
}

func (d *DB) GetNews(num int, days int) (news []News, err error) {
	query := bson.M{}
	if days != 0 {
		duration := time.Duration(-24*days) * time.Hour
		date := time.Now().Add(duration)
		query = bson.M{"date": bson.M{"$gt": date}}
	}
	newsColl := d.session.DB(DB_NAME).C(NEWS_COLL)
	q := newsColl.Find(query).Sort("-date").Limit(num)
	err = q.All(&news)
	return
}

func (d *DB) InsertStats(stats interface{}) error {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	return statsColl.Insert(stats)
}

func (d *DB) InsertBook(book interface{}) error {
	booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
	return booksColl.Insert(book)
}

func (d *DB) RemoveBook(id bson.ObjectId) error {
	booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
	return booksColl.Remove(bson.M{"_id": id})
}

func (d *DB) UpdateBook(id bson.ObjectId, data interface{}) error {
	booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
	return booksColl.Update(bson.M{"_id": id}, bson.M{"$set": data})
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
	booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
	q := booksColl.Find(query).Sort("-_id")
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
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	bookId, err := mr.GetMostVisited(num, statsColl)
	if err != nil {
		return nil, err
	}

	books = make([]Book, num)
	for i, id := range bookId {
		booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
		booksColl.Find(bson.M{"_id": id}).One(&books[i])
		books[i].Id = bson.ObjectId(books[i].Id).Hex()
	}
	return
}

/* Get the most downloaded books
 */
func (d *DB) GetDownloadedBooks(num int) (books []Book, err error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	bookId, err := mr.GetMostDownloaded(num, statsColl)
	if err != nil {
		return nil, err
	}

	books = make([]Book, num)
	for i, id := range bookId {
		booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
		booksColl.Find(bson.M{"_id": id}).One(&books[i])
		books[i].Id = bson.ObjectId(books[i].Id).Hex()
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
	booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
	err := booksColl.Find(bson.M{"_id": id}).One(&book)
	if err != nil {
		return false
	}
	return book.Active
}

func (d *DB) GetFS(prefix string) *mgo.GridFS {
	return d.session.DB(DB_NAME).GridFS(prefix)
}

func (d *DB) GetTags(numTags int) ([]string, error) {
	booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetTags(numTags, booksColl)
}

type Visits struct {
	Date  int64 "_id"
	Count int   "value"
}

func (d *DB) GetHourVisits(start time.Time) ([]Visits, error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetHourVisits(start, statsColl)
}

func (d *DB) GetDayVisits(start time.Time) ([]Visits, error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetDayVisits(start, statsColl)
}

func (d *DB) GetMonthVisits(start time.Time) ([]Visits, error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetMonthVisits(start, statsColl)
}

func (d *DB) GetHourDownloads(start time.Time) ([]Visits, error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetHourDownloads(start, statsColl)
}

func (d *DB) GetDayDownloads(start time.Time) ([]Visits, error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetDayDowloads(start, statsColl)
}

func (d *DB) GetMonthDownloads(start time.Time) ([]Visits, error) {
	statsColl := d.session.DB(DB_NAME).C(STATS_COLL)
	mr := NewMR(d.session.DB(DB_NAME))
	return mr.GetMonthDowloads(start, statsColl)
}
