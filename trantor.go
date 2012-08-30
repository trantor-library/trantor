package main

import (
	"crypto/md5"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
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
		id := bson.ObjectIdHex(r.URL.Path[len("/book/"):])
		books, _, err := GetBook(coll, bson.M{"_id": id})
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

func indexHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data indexData

		/* get the tags */
		tags, err := GetTags(coll)
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
		data.Books, data.Count, _ = GetBook(coll, bson.M{}, 6)
		loadTemplate(w, "index", data)
	}
}

func main() {
	session, err := mgo.Dial(DB_IP)
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
	http.HandleFunc("/read/", readHandler(coll))
	http.HandleFunc("/content/", contentHandler(coll))
	http.HandleFunc("/readnew/", readHandler(newColl))
	http.HandleFunc("/contentnew/", contentHandler(newColl))
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
