package main

import log "github.com/cihub/seelog"

import (
	"git.gitorious.org/go-pkg/epubgo.git"
	"net/http"
	"os"
)

func main() {
	db := initDB()
	defer db.Close()

	for _, file := range os.Args[1:len(os.Args)] {
		uploadEpub(file, db)
	}
}

func uploadEpub(filename string, db *DB) {
	epub, err := epubgo.Open(filename)
	if err != nil {
		log.Error("Not valid epub '", filename, "': ", err)
		return
	}
	defer epub.Close()

	book := parseFile(epub, db)
	title, _ := book["title"].(string)
	_, numTitleFound, _ := db.GetBooks(buildQuery("title:"+title), 1)
	if numTitleFound == 0 {
		book["active"] = true
	}

	file, _ := os.Open(filename)
	defer file.Close()
	id, size, err := StoreNewFile(title+".epub", file, db)
	if err != nil {
		log.Error("Error storing book (", title, "): ", err)
		return
	}

	book["filename"] = id
	book["filenamesize"] = size
	err = db.InsertBook(book)
	if err != nil {
		log.Error("Error storing metadata (", title, "): ", err)
		return
	}
	log.Info("File uploaded: ", filename)
}

type Status struct {
	Upload bool
	Stats  bool
	Search string
}

func GetStatus(h handler) Status {
	return Status{}
}

func loadTemplate(w http.ResponseWriter, tmpl string, data interface{})    {}
func loadTxtTemplate(w http.ResponseWriter, tmpl string, data interface{}) {}
func notFound(h handler)                                                   {}
