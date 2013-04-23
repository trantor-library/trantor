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

	for _, book := range books {
		if book.Path == "" {
			fmt.Println("don't needed -- ", book.Title)
			continue
		}
		fmt.Println(book.Title)

		path := getPath(book)

		id, err := storeFile(path, book)
		if err != nil {
			fmt.Println("storeFile ================", err)
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"active": false})
			continue
		}

		cover, coverSmall, err := cover(path, book)
		if err != nil {
			fmt.Println("cover ================", err)
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"active": false, "file": id})
			continue
		}

		if cover != "" {
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"cover": cover, "coversmall": coverSmall, "file": id}, bson.M{"path": 1})
		} else {
			fmt.Println("No cover ================", book.Title)
			db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"active": false, "file": id})
		}
	}
}

func getPath(book Book) string {
	if !book.Active {
		return "new/" + book.Path
	}
	return "books/" + book.Path
}

func storeFile(path string, book Book) (bson.ObjectId, error) {
	fs := db.GetFS(FS_BOOKS)

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fw, err := fs.Create(book.Title + ".epub")
	if err != nil {
		return "", err
	}
	defer fw.Close()
	id, _ := fw.Id().(bson.ObjectId)

	_, err = io.Copy(fw, file)
	if err != nil {
		return id, err
	}
	return id, nil
}

func cover(path string, book Book) (bson.ObjectId, bson.ObjectId, error) {
	e, err := epubgo.Open(path)
	if err != nil {
		return "", "", err
	}
	defer e.Close()

	cover, coverSmall := GetCover(e, book.Title)
	return cover, coverSmall, err
}
