package main

import (
	"html/template"
	"net/http"
)

import txt_tmpl "text/template"

type Status struct {
	BaseURL string
	FullURL string
	Search  string
	User    string
	IsAdmin bool
	Notif   []Notification
	Home    bool
	About   bool
	News    bool
	Upload  bool
	Stats   bool
	Help    bool
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

var templates = template.Must(template.ParseFiles(TEMPLATE_PATH+"header.html",
	TEMPLATE_PATH+"footer.html",
	TEMPLATE_PATH+"404.html",
	TEMPLATE_PATH+"index.html",
	TEMPLATE_PATH+"about.html",
	TEMPLATE_PATH+"news.html",
	TEMPLATE_PATH+"edit_news.html",
	TEMPLATE_PATH+"book.html",
	TEMPLATE_PATH+"search.html",
	TEMPLATE_PATH+"upload.html",
	TEMPLATE_PATH+"new.html",
	TEMPLATE_PATH+"read.html",
	TEMPLATE_PATH+"edit.html",
	TEMPLATE_PATH+"settings.html",
	TEMPLATE_PATH+"stats.html",
	TEMPLATE_PATH+"help.html",
))

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

var txt_templates = txt_tmpl.Must(txt_tmpl.ParseFiles(TEMPLATE_PATH+"search_rss.xml",
	TEMPLATE_PATH+"news_rss.xml",
))

func loadTxtTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := txt_templates.ExecuteTemplate(w, tmpl, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
