package main

import (
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
)

type aboutData struct {
	S Status
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	var data aboutData
	data.S = GetStatus(w, r)
	data.S.About = true
	loadTemplate(w, "about", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)
	sess.LogOut()
	sess.Notify("Log out!", "Bye bye "+sess.User, "success")
	sess.Save(w, r)
	http.Redirect(w, r, "/", 307)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		sess := GetSession(r)
		if db.UserValid(user, pass) {
			sess.LogIn(user)
			sess.Notify("Successful login!", "Welcome "+user, "success")
		} else {
			sess.Notify("Invalid login!", "user or password invalid", "error")
		}
		sess.Save(w, r)
	}
	http.Redirect(w, r, r.Referer(), 307)
}

type bookData struct {
	S    Status
	Book Book
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	var data bookData
	data.S = GetStatus(w, r)
	id := bson.ObjectIdHex(r.URL.Path[len("/book/"):])
	books, _, err := db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		http.NotFound(w, r)
		return
	}
	db.IncVisit(id)
	data.Book = books[0]
	loadTemplate(w, "book", data)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	file := r.URL.Path[1:]
	db.IncDownload(file)
	http.ServeFile(w, r, file)
}

func fileHandler(path string) {
	h := http.FileServer(http.Dir(path[1:]))
	http.Handle(path, http.StripPrefix(path, h))
}

type indexData struct {
	S     Status
	Books []Book
	VisitedBooks []Book
	DownloadedBooks []Book
	Count int
	Tags  []string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var data indexData

	/* get the tags */
	tags, err := db.GetTags()
	if err == nil {
		length := len(tags)
		if length > TAGS_DISPLAY {
			length = TAGS_DISPLAY
		}
		data.Tags = make([]string, length)
		for i, tag := range tags {
			if i == TAGS_DISPLAY {
				break /* display only 50 */
			}
			if tag.Subject != "" {
				data.Tags[i] = tag.Subject
			}
		}
	}

	data.S = GetStatus(w, r)
	data.S.Home = true
	data.Books, data.Count, _ = db.GetBooks(bson.M{"active": true}, 6)
	data.VisitedBooks, _ = db.GetVisitedBooks(6)
	data.DownloadedBooks, _ = db.GetDownloadedBooks(6)
	loadTemplate(w, "index", data)
}

func main() {
	db = initDB()
	defer db.Close()

	/* create the needed folders */
	var err error
	_, err = os.Stat(BOOKS_PATH)
	if err != nil {
		os.Mkdir(BOOKS_PATH, os.ModePerm)
	}
	_, err = os.Stat(COVER_PATH)
	if err != nil {
		os.Mkdir(COVER_PATH, os.ModePerm)
	}
	_, err = os.Stat(NEW_PATH)
	if err != nil {
		os.Mkdir(NEW_PATH, os.ModePerm)
	}

	/* set up web handlers */
	http.HandleFunc("/book/", bookHandler)
	http.HandleFunc("/search/", searchHandler)
	http.HandleFunc("/upload/", uploadHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/logout/", logoutHandler)
	http.HandleFunc("/new/", newHandler)
	http.HandleFunc("/store/", storeHandler)
	http.HandleFunc("/read/", readHandler)
	http.HandleFunc("/content/", contentHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/about/", aboutHandler)
	http.HandleFunc("/books/", downloadHandler)
	fileHandler("/img/")
	fileHandler("/cover/")
	fileHandler("/css/")
	fileHandler("/js/")
	http.HandleFunc("/", indexHandler)
	panic(http.ListenAndServe(":"+PORT, nil))
}
