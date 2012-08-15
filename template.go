package main

import (
	"html/template"
	"net/http"
)

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// TODO: when finish devel conver to global:
	var templates = template.Must(template.ParseFiles("header.html", "footer.html", "index.html", "about.html", "book.html", "search.html", "upload.html"))

	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
