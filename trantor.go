package main

import (
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

const (
	IP         = "127.0.0.1"
	DB_NAME    = "trantor"
	BOOKS_COLL = "books"
)

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// TODO: when finish devel conver to global:
	var templates = template.Must(template.ParseFiles("head.html", "foot.html", "front.html", "book.html", "search.html"))

	// TODO: use includes
	err := templates.ExecuteTemplate(w, "head.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "foot.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func bookHandler(coll *mgo.Collection, w http.ResponseWriter, r *http.Request) {
	var book Book
	coll.Find(bson.M{"title": r.URL.Path[len("/book/"):]}).One(&book)
	loadTemplate(w, "book", book)
}


func fileHandler(path string) {
	h := http.FileServer(http.Dir(path[1:]))
	http.Handle(path, http.StripPrefix(path, h))
}

func main() {
	session, err := mgo.Dial(IP)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	coll := session.DB(DB_NAME).C(BOOKS_COLL)
	num, _ := coll.Count()

	http.HandleFunc("/book/", func(w http.ResponseWriter, r *http.Request) { bookHandler(coll, w, r) })
	http.HandleFunc("/search/", func(w http.ResponseWriter, r *http.Request) { searchHandler(coll, w, r) })
	http.HandleFunc("/upload/", func(w http.ResponseWriter, r *http.Request) { uploadHandler(coll, w, r) })
	fileHandler("/img/")
	fileHandler("/cover/")
	fileHandler("/books/")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { loadTemplate(w, "front", num) })
	http.ListenAndServe(":8080", nil)
}
