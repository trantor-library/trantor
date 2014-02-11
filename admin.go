package main

import log "github.com/cihub/seelog"

import (
	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
)

type settingsData struct {
	S Status
}

func settingsHandler(h handler) {
	if h.sess.User == "" {
		notFound(h)
		return
	}
	if h.r.Method == "POST" {
		current_pass := h.r.FormValue("currpass")
		pass1 := h.r.FormValue("password1")
		pass2 := h.r.FormValue("password2")
		switch {
		case !h.db.UserValid(h.sess.User, current_pass):
			h.sess.Notify("Password error!", "The current password given don't match with the user password. Try again", "error")
		case pass1 != pass2:
			h.sess.Notify("Passwords don't match!", "The new password and the confirmation password don't match. Try again", "error")
		default:
			h.db.SetPassword(h.sess.User, pass1)
			h.sess.Notify("Password updated!", "Your new password is correctly set.", "success")
		}
	}

	var data settingsData
	data.S = GetStatus(h)
	loadTemplate(h.w, "settings", data)
}

func deleteHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	var titles []string
	var isNew bool
	ids := strings.Split(mux.Vars(h.r)["ids"], "/")
	for _, idStr := range ids {
		if !bson.IsObjectIdHex(idStr) {
			continue
		}

		id := bson.ObjectIdHex(idStr)
		books, _, err := h.db.GetBooks(bson.M{"_id": id})
		if err != nil {
			h.sess.Notify("Book not found!", "The book with id '"+idStr+"' is not there", "error")
			continue
		}
		book := books[0]
		DeleteBook(book, h.db)
		h.db.RemoveBook(id)

		if !book.Active {
			isNew = true
		}
		titles = append(titles, book.Title)
	}
	if titles != nil {
		h.sess.Notify("Removed books!", "The books "+strings.Join(titles, ", ")+" are completly removed", "success")
	}
	h.sess.Save(h.w, h.r)
	if isNew {
		http.Redirect(h.w, h.r, "/new/", http.StatusFound)
	} else {
		http.Redirect(h.w, h.r, "/", http.StatusFound)
	}
}

func editHandler(h handler) {
	idStr := mux.Vars(h.r)["id"]
	if !h.sess.IsAdmin() || !bson.IsObjectIdHex(idStr) {
		notFound(h)
		return
	}
	id := bson.ObjectIdHex(idStr)
	books, _, err := h.db.GetBooks(bson.M{"_id": id})
	if err != nil {
		notFound(h)
		return
	}

	var data bookData
	data.Book = books[0]
	data.S = GetStatus(h)
	loadTemplate(h.w, "edit", data)
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

func saveHandler(h handler) {
	idStr := mux.Vars(h.r)["id"]
	if !h.sess.IsAdmin() || !bson.IsObjectIdHex(idStr) {
		notFound(h)
		return
	}

	id := bson.ObjectIdHex(idStr)
	title := h.r.FormValue("title")
	publisher := h.r.FormValue("publisher")
	date := h.r.FormValue("date")
	description := h.r.FormValue("description")
	author := cleanEmptyStr(h.r.Form["author"])
	subject := cleanEmptyStr(h.r.Form["subject"])
	lang := cleanEmptyStr(h.r.Form["lang"])
	book := map[string]interface{}{"title": title,
		"publisher":   publisher,
		"date":        date,
		"description": description,
		"author":      author,
		"subject":     subject,
		"lang":        lang}
	book["keywords"] = keywords(book)
	err := h.db.UpdateBook(id, book)
	if err != nil {
		notFound(h)
		return
	}

	h.sess.Notify("Book Modified!", "", "success")
	h.sess.Save(h.w, h.r)
	if h.db.BookActive(id) {
		http.Redirect(h.w, h.r, "/book/"+idStr, http.StatusFound)
	} else {
		http.Redirect(h.w, h.r, "/new/", http.StatusFound)
	}
}

type newBook struct {
	TitleFound  int
	AuthorFound int
	B           Book
}
type newData struct {
	S     Status
	Found int
	Books []newBook
	Page  int
	Next  string
	Prev  string
}

func newHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	err := h.r.ParseForm()
	if err != nil {
		http.Error(h.w, err.Error(), http.StatusInternalServerError)
		return
	}
	page := 0
	if len(h.r.Form["p"]) != 0 {
		page, err = strconv.Atoi(h.r.Form["p"][0])
		if err != nil {
			page = 0
		}
	}
	res, num, _ := h.db.GetNewBooks(NEW_ITEMS_PAGE, page*NEW_ITEMS_PAGE)

	var data newData
	data.S = GetStatus(h)
	data.Found = num
	if num-NEW_ITEMS_PAGE*page < NEW_ITEMS_PAGE {
		data.Books = make([]newBook, num-NEW_ITEMS_PAGE*page)
	} else {
		data.Books = make([]newBook, NEW_ITEMS_PAGE)
	}
	for i, b := range res {
		data.Books[i].B = b
		_, data.Books[i].TitleFound, _ = h.db.GetBooks(buildQuery("title:"+b.Title), 1)
		_, data.Books[i].AuthorFound, _ = h.db.GetBooks(buildQuery("author:"+strings.Join(b.Author, " author:")), 1)
	}
	data.Page = page + 1
	if num > (page+1)*NEW_ITEMS_PAGE {
		data.Next = "/new/?p=" + strconv.Itoa(page+1)
	}
	if page > 0 {
		data.Prev = "/new/?p=" + strconv.Itoa(page-1)
	}
	loadTemplate(h.w, "new", data)
}

func storeHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	var titles []string
	ids := strings.Split(mux.Vars(h.r)["ids"], "/")
	for _, idStr := range ids {
		if !bson.IsObjectIdHex(idStr) {
			continue
		}

		id := bson.ObjectIdHex(idStr)
		books, _, err := h.db.GetBooks(bson.M{"_id": id})
		if err != nil {
			h.sess.Notify("Book not found!", "The book with id '"+idStr+"' is not there", "error")
			continue
		}
		book := books[0]
		if err != nil {
			h.sess.Notify("An error ocurred!", err.Error(), "error")
			log.Error("Error storing book '", book.Title, "': ", err.Error())
			continue
		}
		h.db.UpdateBook(id, bson.M{"active": true})
		titles = append(titles, book.Title)
	}
	if titles != nil {
		h.sess.Notify("Store books!", "The books '"+strings.Join(titles, ", ")+"' are stored for public download", "success")
	}
	h.sess.Save(h.w, h.r)
	http.Redirect(h.w, h.r, "/new/", http.StatusFound)
}
