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
	In bool  // one level in depth
	Out bool // one level out depth
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
	i := activeIndx-1
	for i > 0 && strings.Contains(chapters[i].Link, "#") {
		i--
	}
	if i > 0 {
		prev = chapters[i].Link
	}
	i = activeIndx+1
	for i < len(chapters) && strings.Contains(chapters[i].Link, "#") {
		i++
	}
	if i < len(chapters) {
		next = chapters[i].Link
	}
	return next, prev, chapters
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
			data.Next, data.Prev, data.Chapters = chapterList(e, file, id, base)
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
