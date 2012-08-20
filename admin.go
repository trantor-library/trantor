package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

const (
	PATH = "books/"
)

func deleteHandler(coll *mgo.Collection, url string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		// cutre hack: /delete/ and /delnew/ have the same lenght:
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
		http.Redirect(w, r, url, 307)
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

type newData struct {
	S     Status
	Found int
	Books []Book
}

func newHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		res, _ := GetBook(coll, bson.M{})
		var data newData
		data.S = GetStatus(w, r)
		data.Found = len(res)
		data.Books = res
		loadTemplate(w, "new", data)
	}
}

func storeHandler(newColl, coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		id := bson.ObjectIdHex(r.URL.Path[len("/store/"):])
		books, err := GetBook(newColl, bson.M{"_id": id})
		if err != nil {
			http.NotFound(w, r)
			return
		}
		book := books[0]

		path := PATH + book.Title[:1] + "/" + book.Title + ".epub"
		_, err = os.Stat(path)
		for i := 0; err == nil; i++ {
			path := PATH + book.Title[:1] + "/" + book.Title + "_" + strconv.Itoa(i) + ".epub"
			_, err = os.Stat(path)
		}

		os.Mkdir(PATH+book.Title[:1], os.ModePerm)
		cmd := exec.Command("mv", book.Path, path)
		cmd.Run()
		book.Path = path
		coll.Insert(book)
		newColl.Remove(bson.M{"_id": id})
		sess.Notify("Store book!", "The book '"+book.Title+"' it's stored for public download", "success")
		sess.Save(w, r)
		http.Redirect(w, r, "/new/", 307)
	}
}