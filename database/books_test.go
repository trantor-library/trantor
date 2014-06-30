package database

import "testing"

import (
	"labix.org/v2/mgo/bson"
)

var book = map[string]interface{}{
	"title":  "some title",
	"author": []string{"Alice", "Bob"},
}

func TestAddBook(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddBook(t, db)

	books, num, err := db.GetBooks(bson.M{}, 1, 0)
	if err != nil {
		t.Fatalf("db.GetBooks() return an error: ", err)
	}
	if num < 1 {
		t.Fatalf("db.GetBooks() didn't find any result.")
	}
	if len(books) < 1 {
		t.Fatalf("db.GetBooks() didn't return any result.")
	}
	if books[0].Title != book["title"] {
		t.Errorf("Book title don't match : '", books[0].Title, "' <=> '", book["title"], "'")
	}
}

func tAddBook(t *testing.T, db *DB) {
	err := db.AddBook(book)
	if err != nil {
		t.Errorf("db.AddBook(", book, ") return an error: ", err)
	}
}
