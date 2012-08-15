package main

import (
	"html/template"
	"net/http"
)

const (
	TEMPLATE_DIR = "templates/"
)

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// TODO: when finish devel conver to global:
	var templates = template.Must(template.ParseFiles(TEMPLATE_DIR + "header.html",
							  TEMPLATE_DIR + "footer.html",
							  TEMPLATE_DIR + "index.html",
							  TEMPLATE_DIR + "about.html",
							  TEMPLATE_DIR + "book.html",
							  TEMPLATE_DIR + "search.html",
							  TEMPLATE_DIR + "upload.html"))

	err := templates.ExecuteTemplate(w, tmpl+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
