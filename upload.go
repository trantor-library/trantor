package main

import (
	"os"
	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	"net/http"
)

func storeFile(r *http.Request) error {
	f, header, err := r.FormFile("epub")
	if err != nil {
		return err
	}
	defer f.Close()
	// FIXME: check the name exist
	fw, err := os.Create("new/" + header.Filename)
	if err != nil {
		return err
	}
	defer fw.Close()

	const size = 1024
	var n int = size
	buff := make([]byte, size)
	for n == size {
		n, err = f.Read(buff)
		fw.Write(buff)
	}

	return nil
}

func uploadHandler(coll *mgo.Collection, w http.ResponseWriter, r *http.Request) {
	status := ""
	if r.Method == "POST" {
		err := storeFile(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status = "Upload successful."
	}

	loadTemplate(w, "upload", status)
}
