package main

import (
	txt_tmpl "text/template"

	"git.gitorious.org/trantor/trantor.git/database"
	log "github.com/cihub/seelog"

	"encoding/json"
	"errors"
	"html/template"
	"net/http"
)

type Status struct {
	BaseURL  string
	FullURL  string
	Search   string
	User     string
	IsAdmin  bool
	Notif    []Notification
	Home     bool
	About    bool
	News     bool
	Upload   bool
	Stats    bool
	Help     bool
	Dasboard bool
}

func GetStatus(h handler) Status {
	var s Status
	s.BaseURL = "http://" + h.r.Host
	s.FullURL = s.BaseURL + h.r.RequestURI
	s.User = h.sess.User
	s.IsAdmin = h.sess.IsAdmin()
	s.Notif = h.sess.GetNotif()
	h.sess.Save(h.w, h.r)
	return s
}

var tmpl_html = template.Must(template.ParseFiles(
	TEMPLATE_PATH+"header.html",
	TEMPLATE_PATH+"footer.html",
	TEMPLATE_PATH+"404.html",
	TEMPLATE_PATH+"index.html",
	TEMPLATE_PATH+"about.html",
	TEMPLATE_PATH+"news.html",
	TEMPLATE_PATH+"edit_news.html",
	TEMPLATE_PATH+"book.html",
	TEMPLATE_PATH+"search.html",
	TEMPLATE_PATH+"upload.html",
	TEMPLATE_PATH+"login.html",
	TEMPLATE_PATH+"new.html",
	TEMPLATE_PATH+"read.html",
	TEMPLATE_PATH+"edit.html",
	TEMPLATE_PATH+"dashboard.html",
	TEMPLATE_PATH+"settings.html",
	TEMPLATE_PATH+"stats.html",
	TEMPLATE_PATH+"help.html",
))

var tmpl_rss = txt_tmpl.Must(txt_tmpl.ParseFiles(
	TEMPLATE_PATH+"search.rss",
	TEMPLATE_PATH+"news.rss",
))

func loadTemplate(h handler, tmpl string, data interface{}) {
	var err error
	fmt := h.r.FormValue("fmt")
	switch fmt {
	case "rss":
		err = tmpl_rss.ExecuteTemplate(h.w, tmpl+".rss", data)
	case "json":
		err = loadJson(h.w, tmpl, data)
	default:
		err = tmpl_html.ExecuteTemplate(h.w, tmpl+".html", data)
	}
	if err != nil {
		tmpl_html.ExecuteTemplate(h.w, "404.html", data)
		log.Warn("An error ocurred loading the template ", tmpl, ".", fmt, ": ", err)
	}
}

func loadJson(w http.ResponseWriter, tmpl string, data interface{}) error {
	var res []byte
	var err error
	switch tmpl {
	case "index":
		res, err = indexJson(data)
	case "book":
		res, err = bookJson(data)
	case "news":
		res, err = newsJson(data)
	case "search":
		res, err = searchJson(data)
	}
	if err != nil {
		return err
	}
	_, err = w.Write(res)
	return err
}

func indexJson(data interface{}) ([]byte, error) {
	index, ok := data.(indexData)
	if !ok {
		return nil, errors.New("Data is not valid")
	}

	books := make([]map[string]interface{}, len(index.Books))
	for i, book := range index.Books {
		books[i] = bookJsonRaw(book)
	}
	news := newsJsonRaw(index.News)

	return json.Marshal(map[string]interface{}{
		"title":      "Imperial Library of Trantor",
		"url":        index.S.BaseURL,
		"count":      index.Count,
		"news":       news,
		"tags":       index.Tags,
		"last added": books,
	})
}

func bookJson(data interface{}) ([]byte, error) {
	book, ok := data.(bookData)
	if !ok {
		return nil, errors.New("Data is not valid")
	}

	return json.Marshal(bookJsonRaw(book.Book))
}

func newsJson(data interface{}) ([]byte, error) {
	news, ok := data.(newsData)
	if !ok {
		return nil, errors.New("Data is not valid")
	}

	return json.Marshal(newsJsonRaw(news.News))
}

func newsJsonRaw(news []newsEntry) []map[string]string {
	list := make([]map[string]string, len(news))
	for i, n := range news {
		list[i] = map[string]string{
			"date": n.Date,
			"text": n.Text,
		}
	}
	return list
}

func searchJson(data interface{}) ([]byte, error) {
	search, ok := data.(searchData)
	if !ok {
		return nil, errors.New("Data is not valid")
	}

	books := make([]map[string]interface{}, len(search.Books))
	for i, book := range search.Books {
		books[i] = bookJsonRaw(book)
	}
	return json.Marshal(books)
}

func bookJsonRaw(book database.Book) map[string]interface{} {
	cover := ""
	coverSmall := ""
	if book.Cover {
		cover = "/cover/" + book.Id + "/big/" + book.Title + ".jpg"
		coverSmall = "/cover/" + book.Id + "/small/" + book.Title + ".jpg"
	}
	return map[string]interface{}{
		"id":          book.Id,
		"title":       book.Title,
		"author":      book.Author,
		"contributor": book.Contributor,
		"publisher":   book.Publisher,
		"description": book.Description,
		"subject":     book.Subject,
		"date":        book.Date,
		"lang":        book.Lang,
		"isbn":        book.Isbn,
		"size":        book.FileSize,
		"cover":       cover,
		"cover_small": coverSmall,
		"download":    "/download/" + book.Id + "/" + book.Title + ".epub",
		"read":        "/read/" + book.Id,
	}
}
