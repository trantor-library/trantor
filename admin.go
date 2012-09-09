package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
		books, _, err := GetBook(coll, bson.M{"_id": id})
		if err != nil {
			http.NotFound(w, r)
			return
		}
		book := books[0]
		if book.Cover != "" {
			os.RemoveAll(book.Cover[1:])
		}
		if book.CoverSmall != "" {
			os.RemoveAll(book.CoverSmall[1:])
		}
		os.RemoveAll(book.Path)
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
		books, _, err := GetBook(coll, bson.M{"_id": id})
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

		idStr := r.URL.Path[len("/save/"):]
		id := bson.ObjectIdHex(idStr)
		title := r.FormValue("title")
		publisher := r.FormValue("publisher")
		date := r.FormValue("date")
		description := r.FormValue("description")
		author := cleanEmptyStr(r.Form["author"])
		subject := cleanEmptyStr(r.Form["subject"])
		lang := cleanEmptyStr(r.Form["lang"])
		book := map[string]interface{}{"title": title,
			"publisher":   publisher,
			"date":        date,
			"description": description,
			"author":      author,
			"subject":     subject,
			"lang":        lang}
		book["keywords"] = keywords(book)
		err := coll.Update(bson.M{"_id": id}, bson.M{"$set": book})
		if err != nil {
			http.NotFound(w, r)
			return
		}

		sess.Notify("Book Modified!", "", "success")
		sess.Save(w, r)
		http.Redirect(w, r, "/book/"+idStr, 307)
	}
}

type newData struct {
	S     Status
	Found int
	Books []Book
}

func newHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > len("/new/") {
			http.ServeFile(w, r, r.URL.Path[1:])
			return
		}

		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		res, num, _ := GetBook(coll, bson.M{})
		var data newData
		data.S = GetStatus(w, r)
		data.Found = num
		data.Books = res
		loadTemplate(w, "new", data)
	}
}

func ValidFileName(path string, title string, extension string) string {
	title = strings.Replace(title, "/", "_", -1)
	title = strings.Replace(title, "?", "_", -1)
	file := path + "/" + title + extension
	_, err := os.Stat(file)
	for i := 0; err == nil; i++ {
		file = path + "/" + title + "_" + strconv.Itoa(i) + extension
		_, err = os.Stat(file)
	}
	return file
}

func storeHandler(newColl, coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}

		id := bson.ObjectIdHex(r.URL.Path[len("/store/"):])
		var book bson.M
		err := newColl.Find(bson.M{"_id": id}).One(&book)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		title, _ := book["title"].(string)
		path := ValidFileName(BOOKS_PATH + title[:1], title, ".epub")

		oldPath, _ := book["path"].(string)
		os.Mkdir(BOOKS_PATH+title[:1], os.ModePerm)
		cmd := exec.Command("mv", oldPath, path)
		cmd.Run()
		book["path"] = path
		coll.Insert(book)
		newColl.Remove(bson.M{"_id": id})
		sess.Notify("Store book!", "The book '"+title+"' it's stored for public download", "success")
		sess.Save(w, r)
		http.Redirect(w, r, "/new/", 307)
	}
}
