package main

import (
	"git.gitorious.org/go-pkg/epub.git"
	"labix.org/v2/mgo/bson"
	"net/http"
	"regexp"
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

func parseUrl(url string) (string, string, string) {
	exp, _ := regexp.Compile("^(\\/[^\\/]*\\/)([^\\/]*)\\/?(.*)?$")
	res := exp.FindStringSubmatch(url)
	base := res[1]
	id := res[2]
	file := ""
	if len(res) == 4 {
		file = res[3]
	}
	return base, id, file
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

/* return next and prev urls from document and the list of chapters */
func chapterList(e *epub.Epub, file string, id string, base string) (string, string, []chapter) {
	var chapters []chapter
	prev := ""
	next := ""
	tit := e.Titerator(epub.TITERATOR_NAVMAP)
	defer tit.Close()

	activeIndx := -1
	depth := 0
	for ; tit.Valid(); tit.Next() {
		var c chapter
		c.Label = tit.Label()
		c.Link = genLink(id, base, tit.Link())
		if cleanLink(tit.Link()) == file {
			c.Active = true
			activeIndx = len(chapters)
		}
		c.Depth = tit.Depth()
		for c.Depth > depth {
			c.In = append(c.In, true)
			depth++
		}
		for c.Depth < depth {
			c.Out = append(c.Out, true)
			depth--
		}
		chapters = append(chapters, c)
	}

	/* if is the same chapter check the previous */
	i := activeIndx - 1
	for i >= 0 && strings.Contains(chapters[i].Link, "#") {
		i--
	}
	if i >= 0 {
		prev = chapters[i].Link
	}
	i = activeIndx + 1
	for i < len(chapters) && strings.Contains(chapters[i].Link, "#") {
		i++
	}
	if i < len(chapters) {
		next = chapters[i].Link
	}
	return next, prev, chapters
}

func readHandler(w http.ResponseWriter, r *http.Request) {
	base, id, file := parseUrl(r.URL.Path)
	books, _, err := db.GetBooks(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil || len(books) == 0 {
		http.NotFound(w, r)
		return
	}

	var data readData
	data.Book = books[0]
	var bookPath string
	if !data.Book.Active {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}
		data.Back = "/new/"
		bookPath = NEW_PATH + data.Book.Path
	} else {
		data.Back = "/book/" + id
		bookPath = BOOKS_PATH + data.Book.Path
	}
	e, _ := epub.Open(bookPath, 0)
	defer e.Close()
	if file == "" {
		it := e.Iterator(epub.EITERATOR_LINEAR)
		defer it.Close()
		http.Redirect(w, r, base+id+"/"+it.CurrUrl(), http.StatusTemporaryRedirect)
		return
	}

	data.S = GetStatus(w, r)
	data.Next, data.Prev, data.Chapters = chapterList(e, file, id, base)
	data.Content = genLink(id, "/content/", file)
	loadTemplate(w, "read", data)
}

func contentHandler(w http.ResponseWriter, r *http.Request) {
	_, id, file := parseUrl(r.URL.Path)
	books, _, err := db.GetBooks(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil || len(books) == 0 {
		http.NotFound(w, r)
		return
	}
	book := books[0]
	var bookPath string
	if !book.Active {
		sess := GetSession(r)
		if sess.User == "" {
			http.NotFound(w, r)
			return
		}
		bookPath = NEW_PATH + book.Path
	} else {
		bookPath = BOOKS_PATH + book.Path
	}
	e, _ := epub.Open(bookPath, 0)
	defer e.Close()
	if file == "" {
		http.NotFound(w, r)
		return
	}

	w.Write(e.Data(file))
}
