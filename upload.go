package main

import (
	"log"
	"net/http"
	"os"
)

func storeFiles(r *http.Request) ([]string, error) {
	r.ParseMultipartForm(20000000)
	filesForm := r.MultipartForm.File["epub"]
	paths := make([]string, 0, len(filesForm))
	for _, f := range filesForm {
		log.Println("File uploaded:", f.Filename)
		file, err := f.Open()
		if err != nil {
			return paths, err
		}
		defer file.Close()

		path, err := StoreNewFile(f.Filename, file)
		if err != nil {
			return paths, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

type uploadData struct {
	S Status
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		sess := GetSession(r)
		paths, err := storeFiles(r)
		if err != nil {
			sess.Notify("Problem uploading!", "Some files were not stored. Try again or contact us if it keeps happening", "error")
		}

		uploaded := ""
		for _, path := range paths {
			title, err := ParseFile(path)
			if err != nil {
				os.Remove(NEW_PATH + path)
				sess.Notify("Problem uploading!", "The file '"+path+"' is not a well formed epub", "error")
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
