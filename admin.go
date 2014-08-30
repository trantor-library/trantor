package main

import log "github.com/cihub/seelog"

import (
	"net/http"
	"strconv"
	"strings"

	"git.gitorious.org/trantor/trantor.git/database"
	"github.com/gorilla/mux"
)

func deleteHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	var titles []string
	var isNew bool
	ids := strings.Split(mux.Vars(h.r)["ids"], "/")
	for _, id := range ids {
		book, err := h.db.GetBookId(id)
		if err != nil {
			h.sess.Notify("Book not found!", "The book with id '"+id+"' is not there", "error")
			continue
		}
		h.store.Delete(id)
		h.db.DeleteBook(id)

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
	id := mux.Vars(h.r)["id"]
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}
	book, err := h.db.GetBookId(id)
	if err != nil {
		notFound(h)
		return
	}

	var data bookData
	data.Book = book
	data.S = GetStatus(h)
	loadTemplate(h, "edit", data)
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
	id := mux.Vars(h.r)["id"]
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

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
	err := h.db.UpdateBook(id, book)
	if err != nil {
		notFound(h)
		return
	}

	h.sess.Notify("Book Modified!", "", "success")
	h.sess.Save(h.w, h.r)
	if h.db.IsBookActive(id) {
		http.Redirect(h.w, h.r, "/book/"+id, http.StatusFound)
	} else {
		http.Redirect(h.w, h.r, "/new/", http.StatusFound)
	}
}

type newBook struct {
	TitleFound  int
	AuthorFound int
	B           database.Book
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
		_, data.Books[i].TitleFound, _ = h.db.GetBooks("title:"+b.Title, 1, 0)
		_, data.Books[i].AuthorFound, _ = h.db.GetBooks("author:"+strings.Join(b.Author, " author:"), 1, 0)
	}
	data.Page = page + 1
	if num > (page+1)*NEW_ITEMS_PAGE {
		data.Next = "/new/?p=" + strconv.Itoa(page+1)
	}
	if page > 0 {
		data.Prev = "/new/?p=" + strconv.Itoa(page-1)
	}
	loadTemplate(h, "new", data)
}

func storeHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	var titles []string
	ids := strings.Split(mux.Vars(h.r)["ids"], "/")
	for _, id := range ids {
		if id == "" {
			continue
		}
		book, err := h.db.GetBookId(id)
		if err != nil {
			h.sess.Notify("Book not found!", "The book with id '"+id+"' is not there", "error")
			continue
		}
		if err != nil {
			h.sess.Notify("An error ocurred!", err.Error(), "error")
			log.Error("Error getting book for storing '", book.Title, "': ", err.Error())
			continue
		}
		err = h.db.ActiveBook(id)
		if err != nil {
			h.sess.Notify("An error ocurred!", err.Error(), "error")
			log.Error("Error storing book '", book.Title, "': ", err.Error())
			continue
		}
		titles = append(titles, book.Title)
	}
	if titles != nil {
		h.sess.Notify("Store books!", "The books '"+strings.Join(titles, ", ")+"' are stored for public download", "success")
	}
	h.sess.Save(h.w, h.r)
	http.Redirect(h.w, h.r, "/new/", http.StatusFound)
}
