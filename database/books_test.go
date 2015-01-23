package database

import "testing"

var book = map[string]interface{}{
	"title":  "some title",
	"author": []string{"Alice", "Bob"},
	"id":     "r_m-IOzzIbA6QK5w",
}

func TestAddBook(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddBook(t, db)

	books, num, err := db.GetNewBooks(1, 0)
	if err != nil {
		t.Fatal("db.GetBooks() return an error: ", err)
	}
	if num < 1 {
		t.Fatalf("db.GetBooks() didn't find any result.")
	}
	if len(books) < 1 {
		t.Fatalf("db.GetBooks() didn't return any result.")
	}
	if books[0].Title != book["title"] {
		t.Error("Book title don't match : '", books[0].Title, "' <=> '", book["title"], "'")
	}
}

func TestActiveBook(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddBook(t, db)
	books, _, _ := db.GetNewBooks(1, 0)
	id := books[0].Id

	err := db.ActiveBook(id)
	if err != nil {
		t.Fatal("db.ActiveBook(", id, ") return an error: ", err)
	}

	b, err := db.GetBookId(id)
	if err != nil {
		t.Fatal("db.GetBookId(", id, ") return an error: ", err)
	}
	if b.Author[0] != books[0].Author[0] {
		t.Error("Book author don't match : '", b.Author, "' <=> '", book["author"], "'")
	}
}

func TestFlag(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddBook(t, db)
	id, _ := book["id"].(string)
	db.ActiveBook(id)
	id2 := "tfgrBvd2ps_K4iYt"
	b2 := book
	b2["id"] = id2
	err := db.AddBook(b2)
	if err != nil {
		t.Error("db.AddBook(", book, ") return an error:", err)
	}
	db.ActiveBook(id2)
	id3 := "tfgrBvd2ps_K4iY2"
	b3 := book
	b3["id"] = id3
	err = db.AddBook(b3)
	if err != nil {
		t.Error("db.AddBook(", book, ") return an error:", err)
	}
	db.ActiveBook(id3)

	db.FlagBadQuality(id, "1")
	db.FlagBadQuality(id, "2")
	db.FlagBadQuality(id3, "1")

	b, _ := db.GetBookId(id)
	if b.BadQuality != 2 {
		t.Error("The bad quality flag was not increased")
	}
	b, _ = db.GetBookId(id3)
	if b.BadQuality != 1 {
		t.Error("The bad quality flag was not increased")
	}

	books, _, _ := db.GetBooks("flag:bad_quality", 2, 0)
	if len(books) != 2 {
		t.Fatal("Not the right number of results to the flag search:", len(books))
	}
	if books[0].Id != id {
		t.Error("Search for flag bad_quality is not sort right")
	}
}

func tAddBook(t *testing.T, db *DB) {
	err := db.AddBook(book)
	if err != nil {
		t.Error("db.AddBook(", book, ") return an error:", err)
	}
}
