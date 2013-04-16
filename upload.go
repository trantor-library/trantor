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

func uploadPostHandler(w http.ResponseWriter, r *http.Request) {
	sess := GetSession(r)

	uploaded := ""
	r.ParseMultipartForm(20000000)
	filesForm := r.MultipartForm.File["epub"]
	for _, f := range filesForm {
		log.Println("File uploaded:", f.Filename)
		file, err := f.Open()
		if err != nil {
			sess.Notify("Problem uploading!", "The file '"+f.Filename+"' is not a well formed epub: "+err.Error(), "error")
			continue
		}
		defer file.Close()

		epub, err := openMultipartEpub(file)
		if err != nil {
			sess.Notify("Problem uploading!", "The file '"+f.Filename+"' is not a well formed epub: "+err.Error(), "error")
			continue
		}
		defer epub.Close()

		book := parseFile(epub)
		title, _ := book["title"].(string)
		id, err := StoreNewFile(title, file)
		if err != nil {
			log.Println("Error storing book (", title, "):", err)
			continue
		}

		book["file"] = id
		db.InsertBook(book)
		uploaded += " '" + title + "'"
	}
	if uploaded != "" {
		sess.Notify("Upload successful!", "Added the books:"+uploaded+". Thank you for your contribution", "success")
	}

	uploadHandler(w, r)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
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
