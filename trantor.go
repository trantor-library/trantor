package main

import log "github.com/cihub/seelog"

import (
	"github.com/gorilla/mux"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

type statusData struct {
	S Status
}

func aboutHandler(h handler) {
	var data statusData
	data.S = GetStatus(h)
	data.S.About = true
	loadTemplate(h.w, "about", data)
}

func helpHandler(h handler) {
	var data statusData
	data.S = GetStatus(h)
	data.S.Help = true
	loadTemplate(h.w, "help", data)
}

func logoutHandler(h handler) {
	h.sess.LogOut()
	h.sess.Notify("Log out!", "Bye bye "+h.sess.User, "success")
	h.sess.Save(h.w, h.r)
	log.Info("User ", h.sess.User, " log out")
	http.Redirect(h.w, h.r, "/", http.StatusFound)
}

func loginHandler(h handler) {
	user := h.r.FormValue("user")
	pass := h.r.FormValue("pass")
	if h.db.UserValid(user, pass) {
		log.Info("User ", user, " log in")
		h.sess.LogIn(user)
		h.sess.Notify("Successful login!", "Welcome "+user, "success")
	} else {
		log.Warn("User ", user, " bad user or password")
		h.sess.Notify("Invalid login!", "user or password invalid", "error")
	}
	h.sess.Save(h.w, h.r)
	http.Redirect(h.w, h.r, h.r.Referer(), http.StatusFound)
}

type bookData struct {
	S           Status
	Book        Book
	Description []string
}

func bookHandler(h handler) {
	idStr := mux.Vars(h.r)["id"]
	if !bson.IsObjectIdHex(idStr) {
		notFound(h)
		return
	}

	var data bookData
	data.S = GetStatus(h)
	id := bson.ObjectIdHex(idStr)
	books, _, err := h.db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		notFound(h)
		return
	}
	data.Book = books[0]
	data.Description = strings.Split(data.Book.Description, "\n")
	loadTemplate(h.w, "book", data)
}

func downloadHandler(h handler) {
	idStr := mux.Vars(h.r)["id"]
	if !bson.IsObjectIdHex(idStr) {
		notFound(h)
		return
	}

	id := bson.ObjectIdHex(idStr)
	books, _, err := h.db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		notFound(h)
		return
	}
	book := books[0]

	if !book.Active {
		if !h.sess.IsAdmin() {
			notFound(h)
			return
		}
	}

	fs := h.db.GetFS(FS_BOOKS)
	f, err := fs.OpenId(book.File)
	if err != nil {
		notFound(h)
		return
	}
	defer f.Close()

	headers := h.w.Header()
	headers["Content-Type"] = []string{"application/epub+zip"}
	headers["Content-Disposition"] = []string{"attachment; filename=\"" + f.Name() + "\""}

	io.Copy(h.w, f)
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

func indexHandler(h handler) {
	var data indexData

	data.Tags, _ = h.db.GetTags(TAGS_DISPLAY)
	data.S = GetStatus(h)
	data.S.Home = true
	data.Books, data.Count, _ = h.db.GetBooks(bson.M{"active": true}, 6)
	data.VisitedBooks, _ = h.db.GetVisitedBooks(6)
	data.DownloadedBooks, _ = h.db.GetDownloadedBooks(6)
	data.News = getNews(1, DAYS_NEWS_INDEXPAGE, h.db)
	loadTemplate(h.w, "index", data)
}

func notFound(h handler) {
	var data statusData

	data.S = GetStatus(h)
	h.w.WriteHeader(http.StatusNotFound)
	loadTemplate(h.w, "404", data)
}

func updateLogger() error {
	logger, err := log.LoggerFromConfigAsFile(LOGGER_CONFIG)
	if err != nil {
		return err
	}

	return log.ReplaceLogger(logger)
}

func main() {
	defer log.Flush()
	updateLogger()

	db := initDB()
	defer db.Close()

	InitTasks(db)
	InitStats(db)
	InitUpload(db)

	initRouter(db)
	log.Error(http.ListenAndServe(":"+PORT, nil))
}

func initRouter(db *DB) {
	r := mux.NewRouter()
	var notFoundHandler http.HandlerFunc
	notFoundHandler = GatherStats(notFound, db)
	r.NotFoundHandler = notFoundHandler

	r.HandleFunc("/", GatherStats(indexHandler, db))
	r.HandleFunc("/book/{id:[0-9a-fA-F]+}", GatherStats(bookHandler, db))
	r.HandleFunc("/search/", GatherStats(searchHandler, db))
	r.HandleFunc("/upload/", GatherStats(uploadHandler, db)).Methods("GET")
	r.HandleFunc("/upload/", GatherStats(uploadPostHandler, db)).Methods("POST")
	r.HandleFunc("/login/", GatherStats(loginHandler, db)).Methods("POST")
	r.HandleFunc("/logout/", GatherStats(logoutHandler, db))
	r.HandleFunc("/new/", GatherStats(newHandler, db))
	r.HandleFunc("/store/{ids:([0-9a-fA-F]+/)+}", GatherStats(storeHandler, db))
	r.HandleFunc("/delete/{ids:([0-9a-fA-F]+/)+}", GatherStats(deleteHandler, db))
	r.HandleFunc("/read/{id:[0-9a-fA-F]+}", GatherStats(readStartHandler, db))
	r.HandleFunc("/read/{id:[0-9a-fA-F]+}/{file:.*}", GatherStats(readHandler, db))
	r.HandleFunc("/content/{id:[0-9a-fA-F]+}/{file:.*}", GatherStats(contentHandler, db))
	r.HandleFunc("/edit/{id:[0-9a-fA-F]+}", GatherStats(editHandler, db))
	r.HandleFunc("/save/{id:[0-9a-fA-F]+}", GatherStats(saveHandler, db)).Methods("POST")
	r.HandleFunc("/about/", GatherStats(aboutHandler, db))
	r.HandleFunc("/help/", GatherStats(helpHandler, db))
	r.HandleFunc("/download/{id:[0-9a-fA-F]+}/{epub:.*}", GatherStats(downloadHandler, db))
	r.HandleFunc("/cover/{id:[0-9a-fA-F]+}/{size}/{img:.*}", GatherStats(coverHandler, db))
	r.HandleFunc("/settings/", GatherStats(settingsHandler, db))
	r.HandleFunc("/stats/", GatherStats(statsHandler, db))
	r.HandleFunc("/news/", GatherStats(newsHandler, db))
	r.HandleFunc("/news/edit", GatherStats(editNewsHandler, db)).Methods("GET")
	r.HandleFunc("/news/edit", GatherStats(postNewsHandler, db)).Methods("POST")
	h := http.FileServer(http.Dir(IMG_PATH))
	r.Handle("/img/{img}", http.StripPrefix("/img/", h))
	h = http.FileServer(http.Dir(CSS_PATH))
	r.Handle("/css/{css}", http.StripPrefix("/css/", h))
	h = http.FileServer(http.Dir(JS_PATH))
	r.Handle("/js/{js}", http.StripPrefix("/js/", h))
	http.Handle("/", r)
}
