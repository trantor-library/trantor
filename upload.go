package main

import (
	"os"
	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	"net/http"
)

func storeFile(r *http.Request) {
	f, header, err := r.FormFile("epub")
	if err != nil {
		panic(err) // FIXME
	}
	defer f.Close()
	// FIXME: check the name exist
	fw, err := os.Create("new/" + header.Filename)
	if err != nil {
		panic(err) // FIXME
	}
	defer fw.Close()

	const size = 1024
	var n int = size
	buff := make([]byte, size)
	for n == size {
		n, err = f.Read(buff)
		fw.Write(buff)
	}
}

func uploadHandler(coll *mgo.Collection, w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		storeFile(r)
	}

	loadTemplate(w, "upload", nil)
}
