package main

import (
	"os"
	"strconv"
	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	"net/http"
)

func storePath(name string) string {
	path := "new/" + name
	_, err := os.Stat(path)
	for i := 0; err == nil; i++ {
		path = "new/" + strconv.Itoa(i) + "_" + name
		_, err = os.Stat(path)
	}
	return path
}

func storeFile(r *http.Request) error {
	f, header, err := r.FormFile("epub")
	if err != nil {
		return err
	}
	defer f.Close()

	fw, err := os.Create(storePath(header.Filename))
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
