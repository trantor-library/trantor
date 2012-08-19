package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
)

func deleteHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		id := bson.ObjectIdHex(r.URL.Path[len("/delete/"):])
		books, err := GetBook(coll, bson.M{"_id": id})
		if err != nil {
			http.NotFound(w, r)
			return
		}
		book := books[0]
		os.RemoveAll(book.Path)
		os.RemoveAll(book.Cover[1:])
		os.RemoveAll(book.CoverSmall[1:])
		coll.Remove(bson.M{"_id": id})
		sess.Notify("Removed book!", "The book '"+book.Title+"' it's completly removed", "success")
		sess.Save(w, r)
		http.Redirect(w, r, "/", 307)
	}
}

func editHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}
		id := bson.ObjectIdHex(r.URL.Path[len("/edit/"):])
		books, err := GetBook(coll, bson.M{"_id": id})
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var data bookData
		data.Book = books[0]
		data.S = GetStatus(w, r)
		loadTemplate(w, "edit", data)
	}
}

func cleanEmptyStr(s []string) []string {
	var res []string
	for _, v := range s {
		if v != "" {
			res = append(res, v)
		}
	}
	return res
}

func saveHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		id := bson.ObjectIdHex(r.URL.Path[len("/save/"):])
		title := r.FormValue("title")
		publisher := r.FormValue("publisher")
		date := r.FormValue("date")
		description := r.FormValue("description")
		author := cleanEmptyStr(r.Form["author"])
		subject := cleanEmptyStr(r.Form["subject"])
		lang := cleanEmptyStr(r.Form["lang"])
		err := coll.Update(bson.M{"_id": id}, bson.M{"$set": bson.M{"title": title,
			"publisher":   publisher,
			"date":        date,
			"description": description,
			"author":      author,
			"subject":     subject,
			"lang":        lang}})
		if err != nil {
			http.NotFound(w, r)
			return
		}

		sess.Notify("Book Modified!", "", "success")
		sess.Save(w, r)
		http.Redirect(w, r, "/book/"+title, 307)
	}
}

func newHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}
	}
}
