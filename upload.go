package main

import log "github.com/cihub/seelog"

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"io/ioutil"
	"mime/multipart"
	"strings"
)

func InitUpload(database *DB) {
	uploadChannel = make(chan uploadRequest, CHAN_SIZE)
	go uploadWorker(database)
}

var uploadChannel chan uploadRequest

type uploadRequest struct {
	file     multipart.File
	filename string
}

func uploadWorker(database *DB) {
	db := database.Copy()
	defer db.Close()

	for req := range uploadChannel {
		processFile(req, db)
	}
}

func processFile(req uploadRequest, db *DB) {
	defer req.file.Close()

	epub, err := openMultipartEpub(req.file)
	if err != nil {
		log.Warn("Not valid epub uploaded file", req.filename, ":", err)
		return
	}
	defer epub.Close()

	book := parseFile(epub, db)
	title, _ := book["title"].(string)
	req.file.Seek(0, 0)
	id, size, err := StoreNewFile(title+".epub", req.file, db)
	if err != nil {
		log.Error("Error storing book (", title, "):", err)
		return
	}

	book["file"] = id
	book["filesize"] = size
	err = db.InsertBook(book)
	if err != nil {
		log.Error("Error storing metadata (", title, "):", err)
		return
	}
	log.Info("File uploaded:", req.filename)
}

func uploadPostHandler(h handler) {
	problem := false

	h.r.ParseMultipartForm(20000000)
	filesForm := h.r.MultipartForm.File["epub"]
	for _, f := range filesForm {
		file, err := f.Open()
		if err != nil {
			log.Error("Can not open uploaded file", f.Filename, ":", err)
			h.sess.Notify("Upload problem!", "There was a problem with book "+f.Filename, "error")
			problem = true
			continue
		}
		uploadChannel <- uploadRequest{file, f.Filename}
	}

	if !problem {
		if len(filesForm) > 0 {
			h.sess.Notify("Upload successful!", "Thank you for your contribution", "success")
		} else {
			h.sess.Notify("Upload problem!", "No books where uploaded.", "error")
		}
	}
	uploadHandler(h)
}

func uploadHandler(h handler) {
	var data uploadData
	data.S = GetStatus(h)
	data.S.Upload = true
	loadTemplate(h.w, "upload", data)
}

type uploadData struct {
	S Status
}

func openMultipartEpub(file multipart.File) (*epubgo.Epub, error) {
	buff, _ := ioutil.ReadAll(file)
	reader := bytes.NewReader(buff)
	return epubgo.Load(reader, int64(len(buff)))
}

func parseFile(epub *epubgo.Epub, db *DB) map[string]interface{} {
	book := map[string]interface{}{}
	for _, m := range epub.MetadataFields() {
		data, err := epub.Metadata(m)
		if err != nil {
			continue
		}
		switch m {
		case "creator":
			book["author"] = parseAuthr(data)
		case "description":
			book[m] = parseDescription(data)
		case "subject":
			book[m] = parseSubject(data)
		case "date":
			book[m] = parseDate(data)
		case "language":
			book["lang"] = data
		case "title", "contributor", "publisher":
			book[m] = cleanStr(strings.Join(data, ", "))
		case "identifier":
			attr, _ := epub.MetadataAttr(m)
			for i, d := range data {
				if attr[i]["scheme"] == "ISBN" {
					book["isbn"] = d
				}
			}
		default:
			book[m] = strings.Join(data, ", ")
		}
	}
	title, _ := book["title"].(string)
	book["file"] = nil
	cover, coverSmall := GetCover(epub, title, db)
	if cover != "" {
		book["cover"] = cover
		book["coversmall"] = coverSmall
	}
	book["keywords"] = keywords(book)
	return book
}
