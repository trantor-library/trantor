package main

import log "github.com/cihub/seelog"

import (
	"crypto/md5"
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
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
		log.Critical(err)
		os.Exit(1)
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
func (d *DB) GetVisitedBooks() (books []Book, err error) {
	visitedColl := d.session.DB(DB_NAME).C(VISITED_COLL)
	bookId, err := GetBooksVisited(visitedColl)
	if err != nil {
		return nil, err
	}

	books = make([]Book, len(bookId))
	for i, id := range bookId {
		booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
		booksColl.Find(bson.M{"_id": id}).One(&books[i])
		books[i].Id = bson.ObjectId(books[i].Id).Hex()
	}
	return
}

func (d *DB) UpdateMostVisited() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(VISITED_COLL)
	return u.UpdateMostBooks("book")
}

/* Get the most downloaded books
 */
func (d *DB) GetDownloadedBooks() (books []Book, err error) {
	downloadedColl := d.session.DB(DB_NAME).C(DOWNLOADED_COLL)
	bookId, err := GetBooksVisited(downloadedColl)
	if err != nil {
		return nil, err
	}

	books = make([]Book, len(bookId))
	for i, id := range bookId {
		booksColl := d.session.DB(DB_NAME).C(BOOKS_COLL)
		booksColl.Find(bson.M{"_id": id}).One(&books[i])
		books[i].Id = bson.ObjectId(books[i].Id).Hex()
	}
	return
}

func (d *DB) UpdateDownloadedBooks() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(DOWNLOADED_COLL)
	return u.UpdateMostBooks("download")
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

func (d *DB) GetTags() ([]string, error) {
	tagsColl := d.session.DB(DB_NAME).C(TAGS_COLL)
	return GetTags(tagsColl)
}

func (d *DB) UpdateTags() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(BOOKS_COLL)
	u.dst = d.session.DB(DB_NAME).C(TAGS_COLL)
	return u.UpdateTags()
}

type VisitType int

const (
	hourly_visits = iota
	daily_visits
	monthly_visits
	hourly_downloads
	daily_downloads
	monthly_downloads
)

type Visits struct {
	Date  time.Time "date"
	Count int       "count"
}

func (d *DB) GetVisits(visitType VisitType) ([]Visits, error) {
	var coll *mgo.Collection
	switch visitType {
	case hourly_visits:
		coll = d.session.DB(DB_NAME).C(HOURLY_VISITS_COLL)
	case daily_visits:
		coll = d.session.DB(DB_NAME).C(DAILY_VISITS_COLL)
	case monthly_visits:
		coll = d.session.DB(DB_NAME).C(MONTHLY_VISITS_COLL)
	case hourly_downloads:
		coll = d.session.DB(DB_NAME).C(HOURLY_DOWNLOADS_COLL)
	case daily_downloads:
		coll = d.session.DB(DB_NAME).C(DAILY_DOWNLOADS_COLL)
	case monthly_downloads:
		coll = d.session.DB(DB_NAME).C(MONTHLY_DOWNLOADS_COLL)
	default:
		return nil, errors.New("Not valid VisitType")
	}
	return GetVisits(coll)
}

func (d *DB) UpdateHourVisits() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(HOURLY_VISITS_COLL)
	return u.UpdateHourVisits()
}

func (d *DB) UpdateDayVisits() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(DAILY_VISITS_COLL)
	return u.UpdateDayVisits()
}

func (d *DB) UpdateMonthVisits() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(MONTHLY_VISITS_COLL)
	return u.UpdateMonthVisits()
}

func (d *DB) UpdateHourDownloads() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(HOURLY_DOWNLOADS_COLL)
	return u.UpdateHourDownloads()
}

func (d *DB) UpdateDayDownloads() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(DAILY_DOWNLOADS_COLL)
	return u.UpdateDayDownloads()
}

func (d *DB) UpdateMonthDownloads() error {
	var u DBUpdate
	u.src = d.session.DB(DB_NAME).C(STATS_COLL)
	u.dst = d.session.DB(DB_NAME).C(MONTHLY_DOWNLOADS_COLL)
	return u.UpdateMonthDownloads()
}
