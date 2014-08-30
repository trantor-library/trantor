package main

import (
	"net/http"

	"git.gitorious.org/trantor/trantor.git/database"
)

type newsData struct {
	S    Status
	News []newsEntry
}

type newsEntry struct {
	Date string
	Text string
}

func newsHandler(h handler) {
	err := h.r.ParseForm()
	if err != nil {
		http.Error(h.w, err.Error(), http.StatusInternalServerError)
		return
	}

	var data newsData
	data.S = GetStatus(h)
	data.S.News = true
	data.News = getNews(NUM_NEWS, 0, h.db)

	loadTemplate(h, "news", data)
}

func editNewsHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	var data statusData
	data.S = GetStatus(h)
	data.S.News = true
	loadTemplate(h, "edit_news", data)
}

func postNewsHandler(h handler) {
	if !h.sess.IsAdmin() {
		notFound(h)
		return
	}

	text := h.r.FormValue("text")
	h.db.AddNews(text)
	http.Redirect(h.w, h.r, "/news/", http.StatusFound)
}

func getNews(num int, days int, db *database.DB) []newsEntry {
	dbnews, _ := db.GetNews(num, days)
	news := make([]newsEntry, len(dbnews))
	for i, n := range dbnews {
		news[i].Text = n.Text
		news[i].Date = n.Date.Format("Jan 2, 2006")
	}
	return news
}
