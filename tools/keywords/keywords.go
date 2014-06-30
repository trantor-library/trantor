package main

import (
	"fmt"
	"git.gitorious.org/trantor/trantor.git/database"
	"gopkgs.com/unidecode.v1"
	"labix.org/v2/mgo/bson"
	"strings"
	"unicode"
)

func main() {
	db := database.Init(DB_IP, DB_NAME)
	defer db.Close()

	books, _, err := db.GetBooks(bson.M{}, 0, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, b := range books {
		fmt.Println(b.Title)
		book := map[string]interface{}{
			"title":     b.Title,
			"author":    b.Author,
			"publisher": b.Publisher,
			"subject":   b.Subject,
		}
		k := keywords(book)
		book = map[string]interface{}{"keywords": k}
		id := bson.ObjectIdHex(b.Id)
		err := db.UpdateBook(id, book)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func keywords(b map[string]interface{}) (k []string) {
	title, _ := b["title"].(string)
	k = tokens(title)
	author, _ := b["author"].([]string)
	for _, a := range author {
		k = append(k, tokens(a)...)
	}
	publisher, _ := b["publisher"].(string)
	k = append(k, tokens(publisher)...)
	subject, _ := b["subject"].([]string)
	for _, s := range subject {
		k = append(k, tokens(s)...)
	}
	return
}

func tokens(str string) []string {
	str = unidecode.Unidecode(str)
	str = strings.ToLower(str)
	f := func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsPunct(r) || unicode.IsSpace(r)
	}
	return strings.FieldsFunc(str, f)
}
