package main

import (
	"crypto/md5"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"sort"
)

const (
	IP             = "127.0.0.1"
	DB_NAME        = "trantor"
	BOOKS_COLL     = "books"
	NEW_BOOKS_COLL = "new"
	USERS_COLL     = "users"
	PASS_SALT      = "ImperialLibSalt"
)

type aboutData struct {
	S Status
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	var data aboutData
	data.S = GetStatus(w, r)
	loadTemplate(w, "about", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)
	sess.LogOut()
	sess.Notify("Log out!", "Bye bye "+sess.User, "success")
	sess.Save(w, r)
	http.Redirect(w, r, "/", 307)
}

func loginHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			user := r.FormValue("user")
			pass := r.FormValue("pass")
			h := md5.New()
			hash := h.Sum(([]byte)(PASS_SALT + pass))
			n, _ := coll.Find(bson.M{"user": user, "pass": hash}).Count()
			sess := GetSession(r)
			if n != 0 {
				sess.LogIn(user)
				sess.Notify("Successful login!", "Welcome "+user, "success")
			} else {
				sess.Notify("Invalid login!", "user or password invalid", "error")
			}
			sess.Save(w, r)
		}
		http.Redirect(w, r, r.Referer(), 307)
	}
}

type bookData struct {
	S    Status
	Book Book
}

func bookHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data bookData
		data.S = GetStatus(w, r)
		books, err := GetBook(coll, bson.M{"title": r.URL.Path[len("/book/"):]})
		if err != nil || len(books) == 0 {
			http.NotFound(w, r)
			return
		}
		data.Book = books[0]
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
	Tags  []string
}

type tagsList []struct {
	Subject string "_id"
	Count int "value"
}
func (t tagsList) Len() int {
	return len(t)
}
func (t tagsList) Less(i, j int) bool {
	return t[i].Count > t[j].Count
}
func (t tagsList) Swap(i, j int) {
	aux := t[i]
	t[i] = t[j]
	t[j] = aux
}

func indexHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data indexData

		/* get the tags */
		// TODO: cache the tags
		var mr mgo.MapReduce
		mr.Map = "function() { " +
			"this.subject.forEach(function(s) { emit(s, 1); });" +
		"}"
		mr.Reduce = "function(tag, vals) { " +
			"var count = 0;" +
			"vals.forEach(function() { count += 1; });" +
			"return count;" +
		"}"
		var result tagsList
		_, err := coll.Find(nil).MapReduce(&mr, &result)
		if err == nil {
			sort.Sort(result)
			data.Tags = make([]string, len(result))
			for i, tag := range result {
				if i == 50 {
					break /* display only 50 */
				}
				if tag.Subject != "" {
					data.Tags[i] = tag.Subject
				}
			}
		}

		data.S = GetStatus(w, r)
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
	newColl := session.DB(DB_NAME).C(NEW_BOOKS_COLL)

	http.HandleFunc("/book/", bookHandler(coll))
	http.HandleFunc("/search/", searchHandler(coll))
	http.HandleFunc("/upload/", uploadHandler(newColl))
	http.HandleFunc("/login/", loginHandler(userColl))
	http.HandleFunc("/logout/", logoutHandler)
	http.HandleFunc("/new/", newHandler(newColl))
	http.HandleFunc("/delnew/", deleteHandler(newColl, "/new/"))
	http.HandleFunc("/store/", storeHandler(newColl, coll))
	http.HandleFunc("/edit/", editHandler(coll))
	http.HandleFunc("/save/", saveHandler(coll))
	http.HandleFunc("/delete/", deleteHandler(coll, "/"))
	http.HandleFunc("/about/", aboutHandler)
	fileHandler("/img/")
	fileHandler("/cover/")
	fileHandler("/books/")
	fileHandler("/css/")
	fileHandler("/js/")
	http.HandleFunc("/", indexHandler(coll))
	http.ListenAndServe(":8080", nil)
}
