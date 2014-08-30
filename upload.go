package main

import (
	log "github.com/cihub/seelog"

	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io/ioutil"
	"mime/multipart"
	"regexp"
	"strings"

	"git.gitorious.org/go-pkg/epubgo.git"
	"git.gitorious.org/trantor/trantor.git/database"
	"git.gitorious.org/trantor/trantor.git/storage"
)

func InitUpload(database *database.DB, store *storage.Store) {
	uploadChannel = make(chan uploadRequest, CHAN_SIZE)
	go uploadWorker(database, store)
}

var uploadChannel chan uploadRequest

type uploadRequest struct {
	file     multipart.File
	filename string
}

func uploadWorker(database *database.DB, store *storage.Store) {
	db := database.Copy()
	defer db.Close()

	for req := range uploadChannel {
		processFile(req, db, store)
	}
}

func processFile(req uploadRequest, db *database.DB, store *storage.Store) {
	defer req.file.Close()

	epub, err := openMultipartEpub(req.file)
	if err != nil {
		log.Warn("Not valid epub uploaded file ", req.filename, ": ", err)
		return
	}
	defer epub.Close()

	book, id := parseFile(epub, store)
	req.file.Seek(0, 0)
	size, err := store.Store(id, req.file, EPUB_FILE)
	if err != nil {
		log.Error("Error storing book (", id, "): ", err)
		return
	}

	book["filesize"] = size
	err = db.AddBook(book)
	if err != nil {
		log.Error("Error storing metadata (", id, "): ", err)
		return
	}
	log.Info("File uploaded: ", req.filename)
}

func uploadPostHandler(h handler) {
	problem := false

	h.r.ParseMultipartForm(20000000)
	filesForm := h.r.MultipartForm.File["epub"]
	for _, f := range filesForm {
		file, err := f.Open()
		if err != nil {
			log.Error("Can not open uploaded file ", f.Filename, ": ", err)
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
	loadTemplate(h, "upload", data)
}

type uploadData struct {
	S Status
}

func openMultipartEpub(file multipart.File) (*epubgo.Epub, error) {
	buff, _ := ioutil.ReadAll(file)
	reader := bytes.NewReader(buff)
	return epubgo.Load(reader, int64(len(buff)))
}

func parseFile(epub *epubgo.Epub, store *storage.Store) (metadata map[string]interface{}, id string) {
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

	id = genId()
	book["id"] = id //TODO
	book["cover"] = GetCover(epub, id, store)
	return book, id
}

func genId() string {
	b := make([]byte, 12)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func cleanStr(str string) string {
	str = strings.Replace(str, "&#39;", "'", -1)
	exp, _ := regexp.Compile("&[^;]*;")
	str = exp.ReplaceAllString(str, "")
	exp, _ = regexp.Compile("[ ,]*$")
	str = exp.ReplaceAllString(str, "")
	return str
}

func parseAuthr(creator []string) []string {
	exp1, _ := regexp.Compile("^(.*\\( *([^\\)]*) *\\))*$")
	exp2, _ := regexp.Compile("^[^:]*: *(.*)$")
	res := make([]string, len(creator))
	for i, s := range creator {
		auth := exp1.FindStringSubmatch(s)
		if auth != nil {
			res[i] = cleanStr(strings.Join(auth[2:], ", "))
		} else {
			auth := exp2.FindStringSubmatch(s)
			if auth != nil {
				res[i] = cleanStr(auth[1])
			} else {
				res[i] = cleanStr(s)
			}
		}
	}
	return res
}

func parseDescription(description []string) string {
	str := cleanStr(strings.Join(description, "\n"))
	str = strings.Replace(str, "</p>", "\n", -1)
	exp, _ := regexp.Compile("<[^>]*>")
	str = exp.ReplaceAllString(str, "")
	str = strings.Replace(str, "&amp;", "&", -1)
	str = strings.Replace(str, "&lt;", "<", -1)
	str = strings.Replace(str, "&gt;", ">", -1)
	str = strings.Replace(str, "\\n", "\n", -1)
	return str
}

func parseSubject(subject []string) []string {
	var res []string
	for _, s := range subject {
		res = append(res, strings.Split(s, " / ")...)
	}
	return res
}

func parseDate(date []string) string {
	if len(date) == 0 {
		return ""
	}
	return strings.Replace(date[0], "Unspecified: ", "", -1)
}
