package main

import (
	"html/template"
	"git.gitorious.org/go-pkg/epub.git"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"regexp"
	"strings"
)

type readData struct {
	S    Status
	Book Book
	Txt  template.HTML
	Next  string
	Prev  string
}

func parseUrl(url string) (string, string, string) {
	exp, _ := regexp.Compile("^\\/read\\/([^\\/]*)\\/?(.*\\.([^\\.]*))?$")
	res := exp.FindStringSubmatch(url)
	title := res[1]
	file := ""
	ext := ""
	if len(res) == 4 {
		file = res[2]
		ext = res[3]
	}
	return title, file, ext
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

/* return next and prev urls from document */
func nextPrev(e *epub.Epub, file string, title string) (string, string) {
	it := e.Iterator(epub.EITERATOR_LINEAR)
	defer it.Close()

	prev := ""
	next := ""
	for it.CurrUrl() != file {
		prev = it.CurrUrl()
		_, err := it.Next()
		if err != nil {
			break
		}
	}
	if it.CurrUrl() == file {
		_, err := it.Next()
		if err == nil {
			next = it.CurrUrl()
		}
	}
	if prev != "" {
		prev = "/read/" + title + "/" + prev
	}
	if next != "" {
		next = "/read/" + title + "/" + next
	}
	return next, prev
}

func readHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		title, file, ext := parseUrl(r.URL.Path)
		books, _, err := GetBook(coll, bson.M{"title": title})
		if err != nil || len(books) == 0 {
			http.NotFound(w, r)
			return
		}
		book := books[0]
		e, _ := epub.Open(book.Path, 0)
		defer e.Close()
		if file == "" {
			it := e.Iterator(epub.EITERATOR_LINEAR)
			defer it.Close()
			http.Redirect(w, r, "/read/" + title + "/" + it.CurrUrl(), 307)
			return
		}

		if ext == "html" || ext == "htm" || ext == "xhtml" || ext == "xml" {
			var data readData
			data.S = GetStatus(w, r)
			data.Book = book
			data.Next, data.Prev = nextPrev(e, file, title)
			page := string(e.Data(file))
			data.Txt = template.HTML(cleanHtml(page))
			loadTemplate(w, "read", data)
		} else {
			w.Write(e.Data(file))
		}
	}
}
