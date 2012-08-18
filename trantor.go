package main

import (
	"crypto/md5"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

const (
	IP         = "127.0.0.1"
	DB_NAME    = "trantor"
	BOOKS_COLL = "books"
	USERS_COLL = "users"
	PASS_SALT  = "ImperialLibSalt"
)

type aboutData struct {
	S Status
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	var data aboutData
	data.S.User = SessionUser(r)
	data.S.About = true
	loadTemplate(w, "about", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	LogOut(w, r)
	http.Redirect(w, r, "/", 307)
}

func loginHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			user := r.FormValue("user")
			pass := r.FormValue("pass")
			h := md5.New()
			hash := h.Sum(([]byte)(PASS_SALT + pass))
			n, _ := coll.Find(bson.M{"user":user, "pass":hash}).Count()
			if n != 0 {
				// TODO: display success
				CreateSession(user, w, r)
			} else {
				// TODO: display error
			}
		}
		http.Redirect(w, r, "/", 307)
	}
}

type bookData struct {
	S     Status
	Book  Book
}

func bookHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data bookData
		data.S.User = SessionUser(r)
		if coll.Find(bson.M{"title": r.URL.Path[len("/book/"):]}).One(&data.Book) != nil {
			http.NotFound(w, r)
			return
		}
		data.Book.Id = bson.ObjectId(data.Book.Id).Hex()
		loadTemplate(w, "book", data)
	}
}

func fileHandler(path string) {
	h := http.FileServer(http.Dir(path[1:]))
	http.Handle(path, http.StripPrefix(path, h))
}

type indexData struct {
	S     Status
	Books []Book
	Count int
}

func indexHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data indexData
		data.S.User = SessionUser(r)
		data.S.Home = true
		data.Count, _ = coll.Count()
		coll.Find(bson.M{}).Sort("-_id").Limit(6).All(&data.Books)
		loadTemplate(w, "index", data)
	}
}

func main() {
	session, err := mgo.Dial(IP)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	coll := session.DB(DB_NAME).C(BOOKS_COLL)
	userColl := session.DB(DB_NAME).C(USERS_COLL)

	http.HandleFunc("/book/", bookHandler(coll))
	http.HandleFunc("/search/", searchHandler(coll))
	http.HandleFunc("/upload/", uploadHandler(coll))
	http.HandleFunc("/login/", loginHandler(userColl))
	http.HandleFunc("/logout/", logoutHandler)
	http.HandleFunc("/delete/", deleteHandler(coll))
	http.HandleFunc("/about/", aboutHandler)
	fileHandler("/img/")
	fileHandler("/cover/")
	fileHandler("/books/")
	fileHandler("/css/")
	fileHandler("/js/")
	http.HandleFunc("/", indexHandler(coll))
	http.ListenAndServe(":8080", nil)
}
