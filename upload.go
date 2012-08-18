package main

import (
	"labix.org/v2/mgo"
	"os"
	"strconv"
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

type uploadData struct {
	S     Status
	Msg   string
}

func uploadHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var data uploadData
		data.S.User = SessionUser(r)
		data.S.Upload = true
		data.Msg = ""
		if r.Method == "POST" {
			err := storeFile(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			data.Msg = "Upload successful."
		}

		loadTemplate(w, "upload", data)
	}
}
