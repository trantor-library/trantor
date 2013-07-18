package main

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"net/http"
)

func main() {
	db = initDB()
	defer db.Close()
	books, _, _ := db.GetNewBooks()

	for _, book := range books {
		fmt.Println(book.Title)
		e, err := OpenBook(book.File)
		if err != nil {
			fmt.Println("================", err)
		}

		cover, coverSmall := GetCover(e, book.Title)
		if cover != "" {
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"cover": cover, "coversmall": coverSmall})
		}
		e.Close()
	}
}

func notFound(w http.ResponseWriter, r *http.Request) {
	// cover.go needs this function to compile
}
