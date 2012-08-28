package main

import (
	"git.gitorious.org/go-pkg/epub.git"
	"html/template"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"regexp"
	"strings"
	"strconv"
)

type chapter struct {
	Label string
	Link  string
	Depth int
	Active bool
}

type readData struct {
	S    Status
	Book Book
	Txt  template.HTML
	Chapters []chapter
	Next string
	Prev string
	Back string
}

func parseUrl(url string) (string, string, string, string) {
	exp, _ := regexp.Compile("^(\\/read[^\\/]*\\/)([^\\/]*)\\/?(.*\\.([^\\.]*))?$")
	res := exp.FindStringSubmatch(url)
	base := res[1]
	id := res[2]
	file := ""
	ext := ""
	if len(res) == 5 {
		file = res[3]
		ext = res[4]
	}
	return base, id, file, ext
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

/* return next and prev urls from document */
func nextPrev(e *epub.Epub, file string, id string, base string) (string, string) {
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
		it.Next()
		next = it.CurrUrl()
	}
	if prev != "" {
		prev = base + id + "/" + prev
	}
	if next != "" {
		next = base + id + "/" + next
	}
	return next, prev
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

func listChapters(e *epub.Epub, file string, id string, base string) []chapter {
	chapters := make([]chapter, 0)
	tit := e.Titerator(epub.TITERATOR_NAVMAP)
	defer tit.Close()
	for ; tit.Valid(); tit.Next() {
		var c chapter
		c.Label = tit.Label()
		c.Link = genLink(id, base, tit.Link())
		c.Depth = tit.Depth()
		c.Active = cleanLink(tit.Link()) == file
		chapters = append(chapters, c)
	}
	return chapters
}

func readHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		base, id, file, ext := parseUrl(r.URL.Path)
		if base == "/readnew/" {
			sess := GetSession(r)
			if sess.User == "" {
				http.NotFound(w, r)
				return
			}
		}
		books, _, err := GetBook(coll, bson.M{"_id": bson.ObjectIdHex(id)})
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
			http.Redirect(w, r, base+id+"/"+it.CurrUrl(), 307)
			return
		}

		if ext == "html" || ext == "htm" || ext == "xhtml" || ext == "xml" {
			var data readData
			data.S = GetStatus(w, r)
			data.Book = book
			data.Chapters = listChapters(e, file, id, base)
			data.Next, data.Prev = nextPrev(e, file, id, base)
			if base == "/readnew/" {
				data.Back = "/new/"
			} else {
				data.Back = "/book/" + id
			}
			page := string(e.Data(file))
			data.Txt = template.HTML(cleanHtml(page))
			loadTemplate(w, "read", data)
		} else {
			w.Write(e.Data(file))
		}
	}
}
