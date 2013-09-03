package main

import (
	"github.com/gorilla/mux"
	"io"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
	"strings"
)

type statusData struct {
	S Status
}

func aboutHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data statusData
	data.S = GetStatus(w, r)
	data.S.About = true
	loadTemplate(w, "about", data)
}

func helpHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data statusData
	data.S = GetStatus(w, r)
	data.S.Help = true
	loadTemplate(w, "help", data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	sess.LogOut()
	sess.Notify("Log out!", "Bye bye "+sess.User, "success")
	sess.Save(w, r)
	log.Println("User", sess.User, "log out")
	http.Redirect(w, r, "/", http.StatusFound)
}

type bookData struct {
	S           Status
	Book        Book
	Description []string
}

func bookHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	idStr := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(idStr) {
		notFound(w, r)
		return
	}

	var data bookData
	data.S = GetStatus(w, r)
	id := bson.ObjectIdHex(idStr)
	books, _, err := db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		notFound(w, r)
		return
	}
	data.Book = books[0]
	data.Description = strings.Split(data.Book.Description, "\n")
	loadTemplate(w, "book", data)
}

func downloadHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	idStr := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(idStr) {
		notFound(w, r)
		return
	}

	id := bson.ObjectIdHex(idStr)
	books, _, err := db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		notFound(w, r)
		return
	}
	book := books[0]

	if !book.Active {
		sess := GetSession(r)
		if !sess.IsAdmin() {
			notFound(w, r)
			return
		}
	}

	fs := db.GetFS(FS_BOOKS)
	f, err := fs.OpenId(book.File)
	if err != nil {
		notFound(w, r)
		return
	}
	defer f.Close()

	headers := w.Header()
	headers["Content-Type"] = []string{"application/epub+zip"}
	headers["Content-Disposition"] = []string{"attachment; filename=\"" + f.Name() + "\""}

	io.Copy(w, f)
}

type indexData struct {
	S               Status
	Books           []Book
	VisitedBooks    []Book
	DownloadedBooks []Book
	Count           int
	Tags            []string
	News            []newsEntry
}

func indexHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data indexData

	data.Tags, _ = db.GetTags(TAGS_DISPLAY)
	data.S = GetStatus(w, r)
	data.S.Home = true
	data.Books, data.Count, _ = db.GetBooks(bson.M{"active": true}, 6)
	data.VisitedBooks, _ = db.GetVisitedBooks(6)
	data.DownloadedBooks, _ = db.GetDownloadedBooks(6)
	data.News = getNews(1, DAYS_NEWS_INDEXPAGE)
	loadTemplate(w, "index", data)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	var data statusData

	data.S = GetStatus(w, r)
	w.WriteHeader(http.StatusNotFound)
	loadTemplate(w, "404", data)
}

func main() {
	db = initDB()
	defer db.Close()

	InitStats()
	InitUpload()

	setUpRouter()
	panic(http.ListenAndServe(":"+PORT, nil))
}

func setUpRouter() {
	r := mux.NewRouter()
	var notFoundHandler http.HandlerFunc
	notFoundHandler = GatherStats(func(w http.ResponseWriter, r *http.Request, sess *Session) { notFound(w, r) })
	r.NotFoundHandler = notFoundHandler

	r.HandleFunc("/", GatherStats(indexHandler))
	r.HandleFunc("/book/{id:[0-9a-fA-F]+}", GatherStats(bookHandler))
	r.HandleFunc("/search/", GatherStats(searchHandler))
	r.HandleFunc("/upload/", GatherStats(uploadHandler)).Methods("GET")
	r.HandleFunc("/upload/", GatherStats(uploadPostHandler)).Methods("POST")
	r.HandleFunc("/login/", GatherStats(loginHandler)).Methods("GET")
	r.HandleFunc("/login/", GatherStats(loginPostHandler)).Methods("POST")
	r.HandleFunc("/create_user/", GatherStats(createUserHandler)).Methods("POST")
	r.HandleFunc("/logout/", GatherStats(logoutHandler))
	r.HandleFunc("/new/", GatherStats(newHandler))
	r.HandleFunc("/store/{ids:([0-9a-fA-F]+/)+}", GatherStats(storeHandler))
	r.HandleFunc("/delete/{ids:([0-9a-fA-F]+/)+}", GatherStats(deleteHandler))
	r.HandleFunc("/read/{id:[0-9a-fA-F]+}", GatherStats(readStartHandler))
	r.HandleFunc("/read/{id:[0-9a-fA-F]+}/{file:.*}", GatherStats(readHandler))
	r.HandleFunc("/content/{id:[0-9a-fA-F]+}/{file:.*}", GatherStats(contentHandler))
	r.HandleFunc("/edit/{id:[0-9a-fA-F]+}", GatherStats(editHandler))
	r.HandleFunc("/save/{id:[0-9a-fA-F]+}", GatherStats(saveHandler)).Methods("POST")
	r.HandleFunc("/about/", GatherStats(aboutHandler))
	r.HandleFunc("/help/", GatherStats(helpHandler))
	r.HandleFunc("/download/{id:[0-9a-fA-F]+}/{epub:.*}", GatherStats(downloadHandler))
	r.HandleFunc("/cover/{id:[0-9a-fA-F]+}/{size}/{img:.*}", coverHandler)
	r.HandleFunc("/settings/", GatherStats(settingsHandler))
	r.HandleFunc("/stats/", GatherStats(statsHandler))
	r.HandleFunc("/news/", GatherStats(newsHandler))
	r.HandleFunc("/news/edit", GatherStats(editNewsHandler)).Methods("GET")
	r.HandleFunc("/news/edit", GatherStats(postNewsHandler)).Methods("POST")
	h := http.FileServer(http.Dir(IMG_PATH))
	r.Handle("/img/{img}", http.StripPrefix("/img/", h))
	h = http.FileServer(http.Dir(CSS_PATH))
	r.Handle("/css/{css}", http.StripPrefix("/css/", h))
	h = http.FileServer(http.Dir(JS_PATH))
	r.Handle("/js/{js}", http.StripPrefix("/js/", h))
	http.Handle("/", r)
}
