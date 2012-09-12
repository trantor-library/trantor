package main

import (
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)
	if sess.User == "" {
		http.NotFound(w, r)
		return
	}

	var titles []string
	var isNew bool
	ids:= strings.Split(r.URL.Path[len("/delete/"):], "/")
	for _, idStr := range ids {
		if idStr == "" {
			continue
		}

		id := bson.ObjectIdHex(idStr)
		books, _, err := db.GetBooks(bson.M{"_id": id})
		if err != nil {
			sess.Notify("Book not found!", "The book with id '"+idStr+"' is not there", "error")
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
		db.RemoveBook(id)

		if ! book.Active {
			isNew = true
		}
		titles = append(titles, book.Title)
	}
	sess.Notify("Removed books!", "The books "+ strings.Join(titles, ", ") +" are completly removed", "success")
	sess.Save(w, r)
	if isNew {
		http.Redirect(w, r, "/new/", 307)
	} else {
		http.Redirect(w, r, "/", 307)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)
	if sess.User == "" {
		http.NotFound(w, r)
		return
	}
	id := bson.ObjectIdHex(r.URL.Path[len("/edit/"):])
	books, _, err := db.GetBooks(bson.M{"_id": id})
	if err != nil {
		http.NotFound(w, r)
		return
	}

	var data bookData
	data.Book = books[0]
	data.S = GetStatus(w, r)
	loadTemplate(w, "edit", data)
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

func saveHandler(w http.ResponseWriter, r *http.Request) {
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
	err := db.UpdateBook(id, book)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	sess.Notify("Book Modified!", "", "success")
	sess.Save(w, r)
	if db.BookActive(id) {
		http.Redirect(w, r, "/book/"+idStr, 307)
	} else {
		http.Redirect(w, r, "/new/", 307)
	}
}

type newBook struct {
	TitleFound int
	AuthorFound int
	B Book
}
type newData struct {
	S     Status
	Found int
	Books []newBook
}

func newHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)
	if sess.User == "" {
		http.NotFound(w, r)
		return
	}

	if len(r.URL.Path) > len("/new/") {
		http.ServeFile(w, r, r.URL.Path[1:])
		return
	}

	res, num, _ := db.GetNewBooks()
	var data newData
	data.S = GetStatus(w, r)
	data.Found = num
	data.Books = make([]newBook, num)
	for i, b := range res {
		data.Books[i].B = b
		_, data.Books[i].TitleFound, _ = db.GetBooks(buildQuery("title:" + b.Title), 1)
		_, data.Books[i].AuthorFound, _ = db.GetBooks(buildQuery("author:" + strings.Join(b.Author, " author:")), 1)
	}
	loadTemplate(w, "new", data)
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

func storeHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)
	if sess.User == "" {
		http.NotFound(w, r)
		return
	}

	var titles []string
	ids := strings.Split(r.URL.Path[len("/store/"):], "/")
	for _, idStr := range ids {
		if idStr == "" {
			continue
		}

		id := bson.ObjectIdHex(idStr)
		books, _, err := db.GetBooks(bson.M{"_id": id})
		if err != nil {
			sess.Notify("Book not found!", "The book with id '"+idStr+"' is not there", "error")
			return
		}
		book := books[0]

		title := book.Title
		path := ValidFileName(BOOKS_PATH + title[:1], title, ".epub")

		oldPath := book.Path
		os.Mkdir(BOOKS_PATH+title[:1], os.ModePerm)
		cmd := exec.Command("mv", oldPath, path)
		cmd.Run()
		db.UpdateBook(id, bson.M{"active": true, "path": path})
		titles = append(titles, book.Title)
	}
	sess.Notify("Store books!", "The books '"+ strings.Join(titles, ", ") +"' are stored for public download", "success")
	sess.Save(w, r)
	http.Redirect(w, r, "/new/", 307)
}
