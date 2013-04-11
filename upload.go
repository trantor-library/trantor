package main

import (
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

func storeFiles(r *http.Request) ([]bson.ObjectId, error) {
	r.ParseMultipartForm(20000000)
	filesForm := r.MultipartForm.File["epub"]
	ids := make([]bson.ObjectId, 0, len(filesForm))
	for _, f := range filesForm {
		log.Println("File uploaded:", f.Filename)
		file, err := f.Open()
		if err != nil {
			return ids, err
		}
		defer file.Close()

		id, err := StoreNewFile(f.Filename, file)
		if err != nil {
			return ids, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

type uploadData struct {
	S Status
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		sess := GetSession(r)
		ids, err := storeFiles(r)
		if err != nil {
			sess.Notify("Problem uploading!", "Some files were not stored. Try again or contact us if it keeps happening", "error")
		}

		uploaded := ""
		for _, id := range ids {
			title, err := ParseFile(id)
			if err != nil {
				DeleteFile(id)
				sess.Notify("Problem uploading!", "The file is not a well formed epub: "+err.Error(), "error")
			} else {
				uploaded = uploaded + " '" + title + "'"
			}
		}
		if uploaded != "" {
			sess.Notify("Upload successful!", "Added the books:"+uploaded+". Thank you for your contribution", "success")
		}
	}

	var data uploadData
	data.S = GetStatus(w, r)
	data.S.Upload = true
	loadTemplate(w, "upload", data)
}
