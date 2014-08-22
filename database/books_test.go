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

func TestUpdateBookKeywords(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddBook(t, db)
	books, _, _ := db.GetNewBooks(1, 0)

	db.UpdateBook(books[0].Id, map[string]interface{}{"title": "Some other title"})
	books, _, _ = db.GetNewBooks(1, 0)
	keywords := books[0].Keywords

	alice := false
	bob := false
	for _, e := range keywords {
		if e == "alice" {
			alice = true
		}
		if e == "bob" {
			bob = true
		}
	}
	if !alice || !bob {
		t.Error("Alce or Bob are not in the keywords:", keywords)
	}
}

func tAddBook(t *testing.T, db *DB) {
	err := db.AddBook(book)
	if err != nil {
		t.Error("db.AddBook(", book, ") return an error: ", err)
	}
}
