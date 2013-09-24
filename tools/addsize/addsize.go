package main

import (
	"fmt"
	"labix.org/v2/mgo/bson"
)

func main() {
	db = initDB()
	defer db.Close()
	books, _, _ := db.GetBooks(bson.M{})

	for _, book := range books {
		size, err := getSize(book.File)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = db.UpdateBook(bson.ObjectIdHex(book.Id), bson.M{"filesize": size})
		if err != nil {
			fmt.Println(err)
		}
	}
}

type file struct {
	Length int
}

func getSize(id bson.ObjectId) (int, error) {
	fs := db.GetFS(FS_BOOKS)
	var f file
	err := fs.Find(bson.M{"_id": id}).One(&f)
	if err != nil {
		return 0, err
	}
	return f.Length, nil
}
