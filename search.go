package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

func buildQuery(q string) bson.M {
	words := strings.Split(q, " ")
	reg := make([]bson.RegEx, len(words))
	for i, w := range words {
		reg[i].Pattern = w
		reg[i].Options = "i"
	}
	return bson.M{"keywords": bson.M{"$all": reg}}
}

type searchData struct {
	Search string
	Books  []Book
}

func searchHandler(coll *mgo.Collection, w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := strings.Join(r.Form["q"], " ")
	var res []Book
	coll.Find(buildQuery(req)).All(&res)
	loadTemplate(w, "search", searchData{req, res})
}
