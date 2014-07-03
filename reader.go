package main

import (
	"git.gitorious.org/go-pkg/epubgo.git"
	"git.gitorious.org/trantor/trantor.git/database"
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
	Book     database.Book
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
		if cleanLink(spine.URL()) == file {
			break
		}
		prev = spine.URL()
		err = spine.Next()
	}
	if err != nil {
		return "", ""
	}

	if prev != "" {
		prev = genLink(id, base, prev)
	}
	if spine.Next() == nil {
		next = genLink(id, base, spine.URL())
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
		c.Link = nav.URL()
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

func readStartHandler(h handler) {
	id := mux.Vars(h.r)["id"]
	e, _ := openReadEpub(h)
	if e == nil {
		notFound(h)
		return
	}
	defer e.Close()

	it, err := e.Spine()
	if err != nil {
		notFound(h)
		return
	}
	http.Redirect(h.w, h.r, "/read/"+id+"/"+it.URL(), http.StatusTemporaryRedirect)
}

func readHandler(h handler) {
	id := mux.Vars(h.r)["id"]
	file := mux.Vars(h.r)["file"]
	e, book := openReadEpub(h)
	if e == nil {
		notFound(h)
		return
	}
	defer e.Close()

	var data readData
	data.S = GetStatus(h)
	data.Book = book
	if !book.Active {
		data.Back = "/new/"
	} else {
		data.Back = "/book/" + id
	}

	data.Next, data.Prev = getNextPrev(e, file, id, "/read/")
	data.Chapters = getChapters(e, file, id, "/read/")
	data.Content = genLink(id, "/content/", file)
	loadTemplate(h.w, "read", data)
}

func openReadEpub(h handler) (*epubgo.Epub, database.Book) {
	var book database.Book
	id := mux.Vars(h.r)["id"]
	if !bson.IsObjectIdHex(id) {
		return nil, book
	}
	book, err := h.db.GetBookId(id)
	if err != nil {
		return nil, book
	}

	if !book.Active {
		if !h.sess.IsAdmin() {
			return nil, book
		}
	}
	e, err := OpenBook(book.File, h.db)
	if err != nil {
		return nil, book
	}
	return e, book
}

func contentHandler(h handler) {
	vars := mux.Vars(h.r)
	id := vars["id"]
	file := vars["file"]
	if file == "" || !bson.IsObjectIdHex(id) {
		notFound(h)
		return
	}

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
	e, err := OpenBook(book.File, h.db)
	if err != nil {
		notFound(h)
		return
	}
	defer e.Close()

	html, err := e.OpenFile(file)
	if err != nil {
		notFound(h)
		return
	}
	defer html.Close()
	io.Copy(h.w, html)
}
