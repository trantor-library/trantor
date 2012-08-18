package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"os"
)

func deleteHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if SessionUser(r) == "" {
			http.NotFound(w, r)
			return
		}

		var id bson.ObjectId = bson.ObjectIdHex(r.URL.Path[len("/delete/"):])
		var book Book
		if coll.Find(bson.M{"_id": id}).One(&book) != nil {
			http.NotFound(w, r)
			return
		}
		os.RemoveAll(book.Path)
		os.RemoveAll(book.Cover[1:])
		os.RemoveAll(book.CoverSmall[1:])
		coll.Remove(bson.M{"_id": id})
		http.Redirect(w, r, "/", 307)
		//TODO: notify deleted
	}
}
