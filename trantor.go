package main

import (
	"github.com/gorilla/mux"
	"io"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
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
	log.Println("User", sess.User, "log out")
	http.Redirect(w, r, "/", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	sess := GetSession(r)
	if db.UserValid(user, pass) {
		log.Println("User", user, "log in")
		sess.LogIn(user)
		sess.Notify("Successful login!", "Welcome "+user, "success")
	} else {
		log.Println("User", user, "bad user or password")
		sess.Notify("Invalid login!", "user or password invalid", "error")
	}
	sess.Save(w, r)
	http.Redirect(w, r, r.Referer(), http.StatusFound)
}

type bookData struct {
	S    Status
	Book Book
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	var data bookData
	data.S = GetStatus(w, r)
	id := bson.ObjectIdHex(mux.Vars(r)["id"])
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
	id := bson.ObjectIdHex(mux.Vars(r)["id"])
	books, _, err := db.GetBooks(bson.M{"_id": id})
	if err != nil || len(books) == 0 {
		http.NotFound(w, r)
		return
	}
	book := books[0]

	if !book.Active {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}
	}

	fs := db.GetFS(FS_BOOKS)
	f, err := fs.OpenId(book.File)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	headers := w.Header()
	headers["Content-Type"] = []string{"application/epub+zip"}
	headers["Content-Disposition"] = []string{"attachment; filename=\"" + f.Name() + "\""}

	io.Copy(w, f)
	db.IncDownload(id)
}

type indexData struct {
	S               Status
	Books           []Book
	VisitedBooks    []Book
	DownloadedBooks []Book
	Count           int
	Tags            []string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var data indexData

	data.Tags, _ = db.GetTags(TAGS_DISPLAY)
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

	setUpRouter()
	panic(http.ListenAndServe(":"+PORT, nil))
}

func setUpRouter() {
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/book/{id:[0-9a-fA-F]+}", bookHandler)
	r.HandleFunc("/search/", searchHandler)
	r.HandleFunc("/upload/", uploadHandler).Methods("GET")
	r.HandleFunc("/upload/", uploadPostHandler).Methods("POST")
	r.HandleFunc("/login/", loginHandler).Methods("POST")
	r.HandleFunc("/logout/", logoutHandler)
	r.HandleFunc("/new/", newHandler)
	r.HandleFunc("/store/{ids:([0-9a-fA-F]+/)+}", storeHandler)
	r.HandleFunc("/delete/{ids:([0-9a-fA-F]+/)+}", deleteHandler)
	r.HandleFunc("/read/{id:[0-9a-fA-F]+}", readStartHandler)
	r.HandleFunc("/read/{id:[0-9a-fA-F]+}/{file:.*}", readHandler)
	r.HandleFunc("/content/{id:[0-9a-fA-F]+}/{file:.*}", contentHandler)
	r.HandleFunc("/edit/{id:[0-9a-fA-F]+}", editHandler)
	r.HandleFunc("/save/{id:[0-9a-fA-F]+}", saveHandler).Methods("POST")
	r.HandleFunc("/about/", aboutHandler)
	r.HandleFunc("/download/{id:[0-9a-fA-F]+}/{epub:.*}", downloadHandler)
	r.HandleFunc("/cover/{id:[0-9a-fA-F]+}/{size}/{img:.*}", coverHandler)
	r.HandleFunc("/settings/", settingsHandler)
	h := http.FileServer(http.Dir(IMG_PATH))
	r.Handle("/img/{img}", http.StripPrefix("/img/", h))
	h = http.FileServer(http.Dir(CSS_PATH))
	r.Handle("/css/{css}", http.StripPrefix("/css/", h))
	h = http.FileServer(http.Dir(JS_PATH))
	r.Handle("/js/{js}", http.StripPrefix("/js/", h))
	http.Handle("/", r)
}
