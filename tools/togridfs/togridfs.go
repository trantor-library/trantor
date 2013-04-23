package main

import (
	"fmt"
	"git.gitorious.org/go-pkg/epubgo.git"
	"io"
	"labix.org/v2/mgo/bson"
	"os"
)

func main() {
	db = initDB()
	defer db.Close()
	books, _, _ := db.GetBooks(bson.M{})
	fs := db.GetFS(FS_BOOKS)

	for _, book := range books {
		if book.Path == "" {
			fmt.Println("don't needed -- ", book.Title)
			continue
		}
		fmt.Println(book.Title)

		path := "books/" + book.Path
		file, err := os.Open(path)
		if err != nil {
			fmt.Println("os.Open ================", err)
			continue
		}
		defer file.Close()

		fw, err := fs.Create(book.Title + ".epub")
		if err != nil {
			fmt.Println("gridfs.Create ================", err)
			continue
		}
		defer fw.Close()

		_, err = io.Copy(fw, file)
		if err != nil {
			fmt.Println("io.Copy ================", err)
			continue
		}
		id, _ := fw.Id().(bson.ObjectId)

		e, err := epubgo.Open(path)
		if err != nil {
			fmt.Println("epubgo.Open ================", err)
			continue
		}
		defer e.Close()

		cover, coverSmall := GetCover(e, book.Title)
		if cover != "" {
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"cover": cover, "coversmall": coverSmall, "file": id}, bson.M{"path": 1})
		} else {
			fmt.Println("No cover ================", book.Title)
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"file": id}, bson.M{"path": 1})
		}
	}
}
