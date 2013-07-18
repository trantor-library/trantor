package main

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
)

func InitUpload() {
	uploadChannel = make(chan uploadRequest, CHAN_SIZE)
	go uploadWorker()
}

var uploadChannel chan uploadRequest

type uploadRequest struct {
	file     multipart.File
	filename string
}

func uploadWorker() {
	for req := range uploadChannel {
		processFile(req)
	}
}

func processFile(req uploadRequest) {
	defer req.file.Close()

	epub, err := openMultipartEpub(req.file)
	if err != nil {
		log.Println("Not valid epub uploaded file", req.filename, ":", err)
		return
	}
	defer epub.Close()

	book := parseFile(epub)
	title, _ := book["title"].(string)
	req.file.Seek(0, 0)
	id, err := StoreNewFile(title+".epub", req.file)
	if err != nil {
		log.Println("Error storing book (", title, "):", err)
		return
	}

	book["file"] = id
	db.InsertBook(book)
	log.Println("File uploaded:", req.filename)
}

func uploadPostHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	problem := false

	r.ParseMultipartForm(20000000)
	filesForm := r.MultipartForm.File["epub"]
	for _, f := range filesForm {
		file, err := f.Open()
		if err != nil {
			log.Println("Can not open uploaded file", f.Filename, ":", err)
			sess.Notify("Upload problem!", "There was a problem with book "+f.Filename, "error")
			problem = true
			continue
		}
		uploadChannel <- uploadRequest{file, f.Filename}
	}

	if !problem {
		if len(filesForm) > 0 {
			sess.Notify("Upload successful!", "Thank you for your contribution", "success")
		} else {
			sess.Notify("Upload problem!", "No books where uploaded.", "error")
		}
	}
	uploadHandler(w, r, sess)
}

func uploadHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data uploadData
	data.S = GetStatus(w, r)
	data.S.Upload = true
	loadTemplate(w, "upload", data)
}

type uploadData struct {
	S Status
}

func openMultipartEpub(file multipart.File) (*epubgo.Epub, error) {
	buff, _ := ioutil.ReadAll(file)
	reader := bytes.NewReader(buff)
	return epubgo.Load(reader, int64(len(buff)))
}

func parseFile(epub *epubgo.Epub) map[string]interface{} {
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
	cover, coverSmall := GetCover(epub, title)
	book["cover"] = cover
	book["coversmall"] = coverSmall
	book["keywords"] = keywords(book)
	return book
}
