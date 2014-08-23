package main

import (
	log "github.com/cihub/seelog"

	"crypto/rand"
	"encoding/base64"
	"os"

	"git.gitorious.org/trantor/trantor.git/storage"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DB_IP      = "127.0.0.1"
	DB_NAME    = "trantor"
	BOOKS_COLL = "books"
	FS_BOOKS   = "fs_books"
	FS_IMGS    = "fs_imgs"

	STORE_PATH       = "store/"
	EPUB_FILE        = "book.epub"
	COVER_FILE       = "cover.jpg"
	COVER_SMALL_FILE = "coverSmall.jpg"

	NUM_WORKERS = 10
)

type Book struct {
	Id          bson.ObjectId `bson:"_id"`
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

func main() {
	db := InitDB(DB_IP)
	defer db.Close()
	store, err := storage.Init(STORE_PATH)
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}

	channel := make(chan Book)
	quit := make(chan bool)
	for i := 0; i < NUM_WORKERS; i++ {
		go worker(channel, quit, db, store)
	}

	booksColl := db.DB(DB_NAME).C(BOOKS_COLL)
	books := booksColl.Find(bson.M{}).Batch(200).Prefetch(0.25).Iter()
	var book Book
	for books.Next(&book) {
		channel <- book
	}
	if err := books.Close(); err != nil {
		log.Critical(err)
	}
	close(channel)

	for i := 0; i < NUM_WORKERS; i++ {
		log.Info("Worker ", i, " has finished")
		<-quit
	}
}

func InitDB(host string) *mgo.Session {
	session, err := mgo.Dial(host)
	if err != nil {
		log.Critical(err)
		os.Exit(1)
	}
	return session
}

func worker(channel chan Book, quit chan bool, database *mgo.Session, store *storage.Store) {
	db := database.Copy()
	defer db.Close()

	fsBooks := db.DB(DB_NAME).GridFS(FS_BOOKS)
	fsImgs := db.DB(DB_NAME).GridFS(FS_IMGS)
	booksColl := db.DB(DB_NAME).C(BOOKS_COLL)

	for book := range channel {
		id := genId()
		log.Info("== Storing book '", book.Title, "' (", id, ") ==")
		cover := true

		process(id, EPUB_FILE, book.File, fsBooks, store)
		err := process(id, COVER_FILE, book.Cover, fsImgs, store)
		if err != nil {
			cover = false
		}
		process(id, COVER_SMALL_FILE, book.CoverSmall, fsImgs, store)

		query := bson.M{"$set": bson.M{"id": id, "cover": cover},
			"$unset": bson.M{"file": "", "coversmall": ""}}
		err = booksColl.UpdateId(book.Id, query)
		if err != nil {
			log.Error("Can no update ", book.Id.Hex())
		}
	}
	quit <- true
}

func process(id string, name string, objId bson.ObjectId, fs *mgo.GridFS, store *storage.Store) error {
	f, err := fs.OpenId(objId)
	if err != nil {
		if name == EPUB_FILE {
			log.Error(id, " - can not open ", objId.Hex())
		}
		return err
	}
	defer f.Close()

	_, err = store.Store(id, f, name)
	if err != nil {
		log.Error("Can not store '", id, "' (", objId, ")")
	}
	return err
}

func genId() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
