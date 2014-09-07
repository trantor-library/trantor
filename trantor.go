package main

import (
	log "github.com/cihub/seelog"

	"io"
	"net/http"
	"os"
	"strings"

	"git.gitorious.org/trantor/trantor.git/database"
	"git.gitorious.org/trantor/trantor.git/storage"
	"github.com/gorilla/mux"
)

type statusData struct {
	S Status
}

func aboutHandler(h handler) {
	var data statusData
	data.S = GetStatus(h)
	data.S.About = true
	loadTemplate(h, "about", data)
}

func helpHandler(h handler) {
	var data statusData
	data.S = GetStatus(h)
	data.S.Help = true
	loadTemplate(h, "help", data)
}

func logoutHandler(h handler) {
	h.sess.LogOut()
	h.sess.Notify("Log out!", "Bye bye "+h.sess.User, "success")
	h.sess.Save(h.w, h.r)
	log.Info("User ", h.sess.User, " log out")
	http.Redirect(h.w, h.r, "/", http.StatusFound)
}

type bookData struct {
	S           Status
	Book        database.Book
	Description []string
}

func bookHandler(h handler) {
	id := mux.Vars(h.r)["id"]
	var data bookData
	data.S = GetStatus(h)
	book, err := h.db.GetBookId(id)
	if err != nil {
		notFound(h)
		return
	}
	data.Book = book
	data.Description = strings.Split(data.Book.Description, "\n")
	loadTemplate(h, "book", data)
}

func downloadHandler(h handler) {
	id := mux.Vars(h.r)["id"]
	book, err := h.db.GetBookId(id)
	if err != nil {
		notFound(h)
		return
	}

	if !book.Active {
		if !h.sess.IsAdmin() {
			notFound(h)
			return
		}
	}

	f, err := h.store.Get(book.Id, EPUB_FILE)
	if err != nil {
		notFound(h)
		return
	}
	defer f.Close()

	headers := h.w.Header()
	headers["Content-Type"] = []string{"application/epub+zip"}
	headers["Content-Disposition"] = []string{"attachment; filename=\"" + book.Title + ".epub\""}

	io.Copy(h.w, f)
}

type indexData struct {
	S               Status
	Books           []database.Book
	VisitedBooks    []database.Book
	DownloadedBooks []database.Book
	Count           int
	Tags            []string
	News            []newsEntry
}

func indexHandler(h handler) {
	var data indexData

	data.Tags, _ = h.db.GetTags()
	data.S = GetStatus(h)
	data.S.Home = true
	data.Books, data.Count, _ = h.db.GetBooks("", BOOKS_FRONT_PAGE, 0)
	data.VisitedBooks, _ = h.db.GetVisitedBooks()
	data.DownloadedBooks, _ = h.db.GetDownloadedBooks()
	data.News = getNews(1, DAYS_NEWS_INDEXPAGE, h.db)
	loadTemplate(h, "index", data)
}

func notFound(h handler) {
	var data statusData

	data.S = GetStatus(h)
	h.w.WriteHeader(http.StatusNotFound)
	loadTemplate(h, "404", data)
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
	err := updateLogger()
	if err != nil {
		log.Error("Error loading the logger xml: ", err)
	}
	log.Info("Start the imperial library of trantor")

	db := database.Init(DB_IP, DB_NAME)
	defer db.Close()

	store, err := storage.Init(STORE_PATH)
	if err != nil {
		log.Critical("Problem initializing store: ", err)
		os.Exit(1)
	}

	InitTasks(db)
	sg := InitStats(db, store)
	InitUpload(db, store)

	initRouter(db, sg)
	log.Error(http.ListenAndServe(":"+PORT, nil))
}

func initRouter(db *database.DB, sg *StatsGatherer) {
	const id_pattern = "[0-9a-zA-Z\\-\\_]{16}"

	r := mux.NewRouter()
	var notFoundHandler http.HandlerFunc
	notFoundHandler = sg.Gather(notFound)
	r.NotFoundHandler = notFoundHandler

	r.HandleFunc("/", sg.Gather(indexHandler))
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, ROBOTS_PATH) })

	r.HandleFunc("/book/{id:"+id_pattern+"}", sg.Gather(bookHandler))
	r.HandleFunc("/search/", sg.Gather(searchHandler))
	r.HandleFunc("/upload/", sg.Gather(uploadHandler)).Methods("GET")
	r.HandleFunc("/upload/", sg.Gather(uploadPostHandler)).Methods("POST")
	r.HandleFunc("/read/{id:"+id_pattern+"}", sg.Gather(readStartHandler))
	r.HandleFunc("/read/{id:"+id_pattern+"}/{file:.*}", sg.Gather(readHandler))
	r.HandleFunc("/content/{id:"+id_pattern+"}/{file:.*}", sg.Gather(contentHandler))
	r.HandleFunc("/about/", sg.Gather(aboutHandler))
	r.HandleFunc("/help/", sg.Gather(helpHandler))
	r.HandleFunc("/download/{id:"+id_pattern+"}/{epub:.*}", sg.Gather(downloadHandler))
	r.HandleFunc("/cover/{id:"+id_pattern+"}/{size}/{img:.*}", sg.Gather(coverHandler))
	r.HandleFunc("/stats/", sg.Gather(statsHandler))

	r.HandleFunc("/login/", sg.Gather(loginHandler)).Methods("GET")
	r.HandleFunc("/login/", sg.Gather(loginPostHandler)).Methods("POST")
	r.HandleFunc("/create_user/", sg.Gather(createUserHandler)).Methods("POST")
	r.HandleFunc("/logout/", sg.Gather(logoutHandler))
	r.HandleFunc("/dashboard/", sg.Gather(dashboardHandler))
	r.HandleFunc("/settings/", sg.Gather(settingsHandler))

	r.HandleFunc("/new/", sg.Gather(newHandler))
	r.HandleFunc("/save/{id:"+id_pattern+"}", sg.Gather(saveHandler)).Methods("POST")
	r.HandleFunc("/edit/{id:"+id_pattern+"}", sg.Gather(editHandler))
	r.HandleFunc("/store/{ids:("+id_pattern+"/)+}", sg.Gather(storeHandler))
	r.HandleFunc("/delete/{ids:("+id_pattern+"/)+}", sg.Gather(deleteHandler))

	r.HandleFunc("/news/", sg.Gather(newsHandler))
	r.HandleFunc("/news/edit", sg.Gather(editNewsHandler)).Methods("GET")
	r.HandleFunc("/news/edit", sg.Gather(postNewsHandler)).Methods("POST")

	h := http.FileServer(http.Dir(IMG_PATH))
	r.Handle("/img/{img}", http.StripPrefix("/img/", h))
	h = http.FileServer(http.Dir(CSS_PATH))
	r.Handle("/css/{css}", http.StripPrefix("/css/", h))
	h = http.FileServer(http.Dir(JS_PATH))
	r.Handle("/js/{js}", http.StripPrefix("/js/", h))
	http.Handle("/", r)
}
