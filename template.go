package main

import log "github.com/cihub/seelog"
import txt_tmpl "text/template"

import "html/template"

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

var tmpl_html = template.Must(template.ParseFiles(TEMPLATE_PATH+"header.html",
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

var tmpl_rss = txt_tmpl.Must(txt_tmpl.ParseFiles(TEMPLATE_PATH+"search.rss",
	TEMPLATE_PATH+"news.rss",
))

func loadTemplate(h handler, tmpl string, data interface{}) {
	var err error
	fmt := h.r.FormValue("fmt")
	if fmt == "rss" {
		err = tmpl_rss.ExecuteTemplate(h.w, tmpl+".rss", data)
	} else {
		err = tmpl_html.ExecuteTemplate(h.w, tmpl+".html", data)
	}
	if err != nil {
		tmpl_html.ExecuteTemplate(h.w, "404.html", data)
		log.Warn("An error ocurred loading the template ", tmpl, ".", fmt, ": ", err)
	}
}
