package main

import (
	"net/http"
)

type newsData struct {
	S    Status
	News []newsEntry
}

type newsEntry struct {
	Date string
	Text string
}

func newsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data newsData
	data.S = GetStatus(w, r)
	data.S.News = true
	data.News = getNews(NUM_NEWS, 0)

	format := r.Form["fmt"]
	if (len(format) > 0) && (format[0] == "rss") {
		loadTxtTemplate(w, "news_rss.xml", data)
	} else {
		loadTemplate(w, "news", data)
	}
}

func editNewsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	if !sess.IsAdmin() {
		notFound(w, r)
		return
	}

	var data statusData
	data.S = GetStatus(w, r)
	data.S.News = true
	loadTemplate(w, "edit_news", data)
}

func postNewsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	if !sess.IsAdmin() {
		notFound(w, r)
		return
	}

	text := r.FormValue("text")
	db.AddNews(text)
	http.Redirect(w, r, "/news/", http.StatusFound)
}

func getNews(num int, days int) []newsEntry {
	dbnews, _ := db.GetNews(num, days)
	news := make([]newsEntry, len(dbnews))
	for i, n := range dbnews {
		news[i].Text = n.Text
		news[i].Date = n.Date.Format("Jan 2, 2006")
	}
	return news
}
