package main

import (
	"net/http"
)

type newsData struct {
	S    Status
	News []news
}

type news struct {
	Date string
	Text string
}

func newsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data newsData
	data.S = GetStatus(w, r)
	data.S.News = true
	newsEntries, _ := db.GetNews(NUM_NEWS)
	data.News = make([]news, len(newsEntries))
	for i, n := range newsEntries {
		data.News[i].Text = n.Text
		data.News[i].Date = n.Date.Format("Jan 31, 2006")
	}
	loadTemplate(w, "news", data)
}

func editNewsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	if sess.User == "" {
		notFound(w, r)
		return
	}

	var data statusData
	data.S = GetStatus(w, r)
	data.S.News = true
	loadTemplate(w, "edit_news", data)
}

func postNewsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	if sess.User == "" {
		notFound(w, r)
		return
	}

	text := r.FormValue("text")
	db.AddNews(text)
	http.Redirect(w, r, "/news/", http.StatusFound)
}
