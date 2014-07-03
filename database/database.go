package database

import log "github.com/cihub/seelog"

import (
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
)

const (
	visited_coll    = "visited"
	downloaded_coll = "downloaded"
	tags_coll       = "tags"
)

type DB struct {
	session *mgo.Session
	name    string
}

func Init(host string, name string) *DB {
	var err error
	db := new(DB)
	db.session, err = mgo.Dial(host)
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}
	db.name = name
	return db
}

func (db *DB) Close() {
	db.session.Close()
}

func (db *DB) Copy() *DB {
	dbCopy := new(DB)
	dbCopy.session = db.session.Copy()
	dbCopy.name = db.name
	return dbCopy
}

func (db *DB) AddBook(book map[string]interface{}) error {
	booksColl := db.session.DB(db.name).C(books_coll)
	return addBook(booksColl, book)
}

func (db *DB) GetBooks(query string, length int, start int) (books []Book, num int, err error) {
	booksColl := db.session.DB(db.name).C(books_coll)
	return getBooks(booksColl, query, length, start)
}

func (db *DB) GetNewBooks(length int, start int) (books []Book, num int, err error) {
	booksColl := db.session.DB(db.name).C(books_coll)
	return getNewBooks(booksColl, length, start)
}

func (db *DB) GetBookId(id string) (Book, error) {
	booksColl := db.session.DB(db.name).C(books_coll)
	return getBookId(booksColl, id)
}

func (db *DB) DeleteBook(id string) error {
	booksColl := db.session.DB(db.name).C(books_coll)
	return deleteBook(booksColl, id)
}

func (db *DB) UpdateBook(id string, data map[string]interface{}) error {
	booksColl := db.session.DB(db.name).C(books_coll)
	return updateBook(booksColl, id, data)
}

func (db *DB) BookActive(id string) bool {
	booksColl := db.session.DB(db.name).C(books_coll)
	return bookActive(booksColl, id)
}

func (db *DB) User(name string) *User {
	userColl := db.session.DB(db.name).C(user_coll)
	return getUser(userColl, name)
}

func (db *DB) AddUser(name string, pass string) error {
	userColl := db.session.DB(db.name).C(user_coll)
	return addUser(userColl, name, pass)
}

func (db *DB) AddNews(text string) error {
	newsColl := db.session.DB(db.name).C(news_coll)
	return addNews(newsColl, text)
}

func (db *DB) GetNews(num int, days int) (news []News, err error) {
	newsColl := db.session.DB(db.name).C(news_coll)
	return getNews(newsColl, num, days)
}

// TODO: split code in files
func (db *DB) AddStats(stats interface{}) error {
	statsColl := db.session.DB(db.name).C(stats_coll)
	return statsColl.Insert(stats)
}

/* Get the most visited books
 */
func (db *DB) GetVisitedBooks() (books []Book, err error) {
	visitedColl := db.session.DB(db.name).C(visited_coll)
	bookId, err := GetBooksVisited(visitedColl)
	if err != nil {
		return nil, err
	}

	books = make([]Book, len(bookId))
	for i, id := range bookId {
		booksColl := db.session.DB(db.name).C(books_coll)
		booksColl.Find(bson.M{"_id": id}).One(&books[i])
		books[i].Id = bson.ObjectId(books[i].Id).Hex()
	}
	return
}

func (db *DB) UpdateMostVisited() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(visited_coll)
	return u.UpdateMostBooks("book")
}

/* Get the most downloaded books
 */
func (db *DB) GetDownloadedBooks() (books []Book, err error) {
	downloadedColl := db.session.DB(db.name).C(downloaded_coll)
	bookId, err := GetBooksVisited(downloadedColl)
	if err != nil {
		return nil, err
	}

	books = make([]Book, len(bookId))
	for i, id := range bookId {
		booksColl := db.session.DB(db.name).C(books_coll)
		booksColl.Find(bson.M{"_id": id}).One(&books[i])
		books[i].Id = bson.ObjectId(books[i].Id).Hex()
	}
	return
}

func (db *DB) UpdateDownloadedBooks() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(downloaded_coll)
	return u.UpdateMostBooks("download")
}

func (db *DB) GetFS(prefix string) *mgo.GridFS {
	return db.session.DB(db.name).GridFS(prefix)
}

func (db *DB) GetTags() ([]string, error) {
	tagsColl := db.session.DB(db.name).C(tags_coll)
	return GetTags(tagsColl)
}

func (db *DB) UpdateTags() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(books_coll)
	u.dst = db.session.DB(db.name).C(tags_coll)
	return u.UpdateTags()
}

func (db *DB) GetVisits(visitType VisitType) ([]Visits, error) {
	var coll *mgo.Collection
	switch visitType {
	case Hourly_visits:
		coll = db.session.DB(db.name).C(hourly_visits_coll)
	case Daily_visits:
		coll = db.session.DB(db.name).C(daily_visits_coll)
	case Monthly_visits:
		coll = db.session.DB(db.name).C(monthly_visits_coll)
	case Hourly_downloads:
		coll = db.session.DB(db.name).C(hourly_downloads_coll)
	case Daily_downloads:
		coll = db.session.DB(db.name).C(daily_downloads_coll)
	case Monthly_downloads:
		coll = db.session.DB(db.name).C(monthly_downloads_coll)
	default:
		return nil, errors.New("Not valid VisitType")
	}
	return GetVisits(coll)
}

func (db *DB) UpdateHourVisits() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(hourly_visits_coll)
	return u.UpdateHourVisits(false)
}

func (db *DB) UpdateDayVisits() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(daily_visits_coll)
	return u.UpdateDayVisits(false)
}

func (db *DB) UpdateMonthVisits() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(monthly_visits_coll)
	return u.UpdateMonthVisits(false)
}

func (db *DB) UpdateHourDownloads() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(hourly_downloads_coll)
	return u.UpdateHourVisits(true)
}

func (db *DB) UpdateDayDownloads() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(daily_downloads_coll)
	return u.UpdateDayVisits(true)
}

func (db *DB) UpdateMonthDownloads() error {
	var u dbUpdate
	u.src = db.session.DB(db.name).C(stats_coll)
	u.dst = db.session.DB(db.name).C(monthly_downloads_coll)
	return u.UpdateMonthVisits(true)
}

// function defined for the tests
func (db *DB) del() {
	defer db.Close()
	db.session.DB(db.name).DropDatabase()
}
