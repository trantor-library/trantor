package main

import (
	"fmt"
	"git.gitorious.org/go-pkg/epub.git"
	"labix.org/v2/mgo/bson"
)

func main() {
	db = initDB()
	defer db.Close()
	books, _, _ := db.GetBooks(bson.M{})

	for _, book := range books {
		fmt.Println(book.Title)
		e, err := epub.Open(BOOKS_PATH+book.Path, 0)
		if err != nil {
			fmt.Println("================", err)
		}

		cover, coverSmall := getCover(e, book.Title)
		if cover != "" {
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"cover": cover, "coversmall": coverSmall})
		}
		e.Close()
	}
}
