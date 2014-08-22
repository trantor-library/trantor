package storage

import "testing"

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

const (
	test_path = "/tmp/store"
	test_book = `
		HARI SELDON-… born in the 11,988th year of the Galactic Era; died
		12,069. The dates are more commonly given in terms of the current
		Foundational Era as - 79 to the year 1 F.E. Born to middle-class
		parents on Helicon, Arcturus sector (where his father, in a legend of
		doubtful authenticity, was a tobacco grower in the hydroponic plants
		of the planet), he early showed amazing ability in mathematics.
		Anecdotes concerning his ability are innumerable, and some are 
		contradictory. At the age of two, he is said to have …`
	test_id = "1234567890abcdef"
)

func TestInit(t *testing.T) {
	st, err := Init(test_path)
	if err != nil {
		t.Fatal("An error ocurred initializing the store =>", err)
	}
	defer st.del()

	info, err := os.Stat(test_path)
	if err != nil {
		t.Fatal("An error ocurred =>", err)
	}
	if !info.Mode().IsDir() {
		t.Errorf(test_path, " is not dir.")
	}

	info, err = os.Stat(test_path + "/a/M")
	if err != nil {
		t.Fatal("An error ocurred =>", err)
	}
	if !info.Mode().IsDir() {
		t.Errorf(test_path, " is not dir.")
	}
}

func TestStore(t *testing.T) {
	st, err := Init(test_path)
	defer st.del()

	_, err = st.Store(test_id, strings.NewReader(test_book), "epub")
	if err != nil {
		t.Fatal("An error ocurred storing the book =>", err)
	}
	book, err := st.Get(test_id, "epub")
	if err != nil {
		t.Fatal("An error ocurred getting the book =>", err)
	}

	content, err := ioutil.ReadAll(book)
	if err != nil {
		t.Fatal("An error ocurred reading the book =>", err)
	}
	if !bytes.Equal(content, []byte(test_book)) {
		t.Error("Not the same content")
	}
}

func TestCreate(t *testing.T) {
	st, err := Init(test_path)
	defer st.del()

	f, err := st.Create(test_id, "img")
	if err != nil {
		t.Fatal("An error ocurred storing the book =>", err)
	}
	io.Copy(f, strings.NewReader(test_book))
	img, err := st.Get(test_id, "img")
	if err != nil {
		t.Fatal("An error ocurred getting the book =>", err)
	}

	content, err := ioutil.ReadAll(img)
	if err != nil {
		t.Fatal("An error ocurred reading the book =>", err)
	}
	if !bytes.Equal(content, []byte(test_book)) {
		t.Error("Not the same content")
	}
}

func TestDelete(t *testing.T) {
	st, err := Init(test_path)
	defer st.del()

	_, err = st.Store(test_id, strings.NewReader(test_book), "epub")
	if err != nil {
		t.Fatal("An error ocurred storing the book =>", err)
	}
	err = st.Delete(test_id)
	if err != nil {
		t.Fatal("An error ocurred deleteing id =>", err)
	}

	_, err = st.Get(test_id, "epub")
	if err == nil {
		t.Fatal("Retrieve book without error.")
	}
}
