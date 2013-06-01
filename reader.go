package main

import (
	"git.gitorious.org/go-pkg/epubgo.git"
	"github.com/gorilla/mux"
	"io"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
)

type chapter struct {
	Label  string
	Link   string
	Depth  int
	Active bool
	In     []bool // one level in depth
	Out    []bool // one level out depth
}

type readData struct {
	S        Status
	Book     Book
	Content  string
	Chapters []chapter
	Next     string
	Prev     string
	Back     string
}

func cleanHtml(html string) string {
	str := strings.Split(html, "<body")
	if len(str) < 2 {
		return html
	}
	str = strings.SplitN(str[1], ">", 2)
	if len(str) < 2 {
		return str[0]
	}

	return "<div " + str[0] + ">" + strings.Split(str[1], "</body>")[0] + "</div>"
}

func genLink(id string, base string, link string) string {
	return base + id + "/" + link
}

func cleanLink(link string) string {
	for i := 0; i < len(link); i++ {
		if link[i] == '%' {
			c, _ := strconv.ParseInt(link[i+1:i+3], 16, 0)
			link = link[:i] + string(c) + link[i+3:]
		}
	}
	return link
}

func getNextPrev(e *epubgo.Epub, file string, id string, base string) (string, string) {
	spine, err := e.Spine()
	if err != nil {
		return "", ""
	}

	prev := ""
	next := ""
	for err == nil {
		if cleanLink(spine.Url()) == file {
			break
		}
		prev = spine.Url()
		err = spine.Next()
	}
	if err != nil {
		return "", ""
	}

	if prev != "" {
		prev = genLink(id, base, prev)
	}
	if spine.Next() == nil {
		next = genLink(id, base, spine.Url())
	}
	return next, prev
}

func getChapters(e *epubgo.Epub, file string, id string, base string) []chapter {
	nav, err := e.Navigation()
	if err != nil {
		return nil
	}
	chapters := listChapters(nav, 0)

	for i, c := range chapters {
		chapters[i].Link = genLink(id, base, c.Link)
		if cleanLink(c.Link) == file {
			chapters[i].Active = true
		}
	}
	return chapters
}

func listChapters(nav *epubgo.NavigationIterator, depth int) []chapter {
	var chapters []chapter
	var err error = nil
	for err == nil {
		var c chapter
		c.Label = nav.Title()
		c.Link = nav.Url()
		c.Depth = depth
		for c.Depth < depth {
			c.Out = append(c.Out, true)
			depth--
		}
		chapters = append(chapters, c)

		if nav.HasChildren() {
			nav.In()
			children := listChapters(nav, depth+1)
			children[0].In = []bool{true}
			children[len(children)-1].Out = []bool{true}
			chapters = append(chapters, children...)
			nav.Out()
		}
		err = nav.Next()
	}
	chapters[0].In = []bool{true}
	chapters[len(chapters)-1].Out = []bool{true}
	return chapters
}

func readStartHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	id := mux.Vars(r)["id"]
	e, _ := openReadEpub(w, r, sess)
	if e == nil {
		notFound(w, r)
		return
	}
	defer e.Close()

	it, err := e.Spine()
	if err != nil {
		notFound(w, r)
		return
	}
	http.Redirect(w, r, "/read/"+id+"/"+it.Url(), http.StatusTemporaryRedirect)
}

func readHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	id := mux.Vars(r)["id"]
	file := mux.Vars(r)["file"]
	e, book := openReadEpub(w, r, sess)
	if e == nil {
		notFound(w, r)
		return
	}
	defer e.Close()

	var data readData
	data.S = GetStatus(w, r)
	data.Book = book
	if !book.Active {
		data.Back = "/new/"
	} else {
		data.Back = "/book/" + id
	}

	data.Next, data.Prev = getNextPrev(e, file, id, "/read/")
	data.Chapters = getChapters(e, file, id, "/read/")
	data.Content = genLink(id, "/content/", file)
	loadTemplate(w, "read", data)
}

func openReadEpub(w http.ResponseWriter, r *http.Request, sess *Session) (*epubgo.Epub, Book) {
	var book Book
	id := mux.Vars(r)["id"]
	if !bson.IsObjectIdHex(id) {
		return nil, book
	}
	books, _, err := db.GetBooks(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil || len(books) == 0 {
		return nil, book
	}

	book = books[0]
	if !book.Active {
		if sess.User == "" {
			return nil, book
		}
	}
	e, err := OpenBook(book.File)
	if err != nil {
		return nil, book
	}
	return e, book
}

func contentHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	vars := mux.Vars(r)
	id := vars["id"]
	file := vars["file"]
	if file == "" || !bson.IsObjectIdHex(id) {
		notFound(w, r)
		return
	}

	books, _, err := db.GetBooks(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil || len(books) == 0 {
		notFound(w, r)
		return
	}
	book := books[0]
	if !book.Active {
		if sess.User == "" {
			notFound(w, r)
			return
		}
	}
	e, err := OpenBook(book.File)
	if err != nil {
		notFound(w, r)
		return
	}
	defer e.Close()

	html, err := e.OpenFile(file)
	if err != nil {
		notFound(w, r)
		return
	}
	defer html.Close()
	io.Copy(w, html)
}
