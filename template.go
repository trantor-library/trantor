package main

import (
	"html/template"
	"net/http"
)

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// TODO: when finish devel conver to global:
	var templates = template.Must(template.ParseFiles("head.html", "foot.html", "front.html", "book.html", "search.html", "upload.html"))

	// TODO: use includes
	err := templates.ExecuteTemplate(w, "head.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = templates.ExecuteTemplate(w, "foot.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
