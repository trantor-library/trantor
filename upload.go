package main

import (
	"fmt"
	"os"
	"git.gitorious.org/go-pkg/epub.git"
	"labix.org/v2/mgo"
)

const (
	IP = "127.0.0.1"
	DB_NAME = "trantor"
	BOOKS_COLL = "books"
	PATH = "./books/"
)


func store(coll *mgo.Collection, path string) {
	var book Book

	e, err := epub.Open(path, 0)
	if err != nil {
		fmt.Println(path)
		panic(err)  // TODO: do something
	}
	defer e.Close()

	// TODO: do it for all metadata
	book.Title = e.Metadata(epub.EPUB_TITLE)
	book.Creator = e.Metadata(epub.EPUB_CREATOR)
	book.Subject = e.Metadata(epub.EPUB_SUBJECT)
	book.Lang = e.Metadata(epub.EPUB_LANG)
	book.Path = path

	coll.Insert(book)
}


func main() {
	session, err := mgo.Dial(IP)
	if err != nil {
		panic(err)  // TODO: do something
	}
	defer session.Close()
	coll := session.DB(DB_NAME).C(BOOKS_COLL)

	f, err := os.Open(PATH)
	if err != nil {
		fmt.Println(PATH)
		panic(err)  // TODO: do something
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		panic(err)  // TODO: do something
	}

	for _, name := range names {
		store(coll, PATH + name)
	}
}
