package main

import (
	"fmt"
	"git.gitorious.org/go-pkg/epubgo.git"
	"labix.org/v2/mgo/bson"
)

func main() {
	db = initDB()
	defer db.Close()
	books, _, _ := db.GetBooks(bson.M{})

	for _, book := range books {
		fmt.Println(book.Title)
		e, err := OpenBook(book.File)
		if err != nil {
			fmt.Println("================", err)
			continue
		}

		updateISBN(e, book)
		updateDescription(e, book)
		e.Close()
	}
}

func updateISBN(e *epubgo.Epub, book Book) {
	attr, err := e.MetadataAttr("identifier")
	if err != nil {
		fmt.Println("isbn ================", err)
		return
	}
	data, err := e.Metadata("identifier")
	if err != nil {
		fmt.Println("isbn ================", err)
		return
	}
	var isbn string
	for i, d := range data {
		if attr[i]["scheme"] == "ISBN" {
			isbn = d
		}
	}

	if isbn != "" {
		db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"isbn": isbn})
	}
}

func updateDescription(e *epubgo.Epub, book Book) {
	descList, err := e.Metadata("description")
	if err != nil {
		fmt.Println("desc ================", err)
		return
	}
	description := parseDescription(descList)
	if len(description) < 10 {
		return
	}

	if len(book.Description) < 10 || book.Description[:10] == description[:10] {
		db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"description": description})
	}
}
