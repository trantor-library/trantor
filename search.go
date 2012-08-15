package main

import (
	"strconv"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
)

const (
	ITEMS_PAGE = 10
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
	Found  int
	Books  []Book
	Page int
	Next string
	Prev string
}

func searchHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		req := strings.Join(r.Form["q"], " ")
		var res []Book
		coll.Find(buildQuery(req)).All(&res)

		page := 0
		if len(r.Form["p"]) != 0 {
			page, err = strconv.Atoi(r.Form["p"][0])
			if err != nil || len(res) < ITEMS_PAGE*page {
				page = 0
			}
		}

		var data searchData
		data.Search = req
		data.Found = len(res)
		if len(res) > ITEMS_PAGE*(page+1) {
			data.Books = res[ITEMS_PAGE*page:ITEMS_PAGE*(page+1)]
		} else {
			data.Books = res[ITEMS_PAGE*page:]
		}
		data.Page = page+1
		if len(res) > (page+1)*ITEMS_PAGE {
			data.Next = "/search/?q=" + req + "&p=" + strconv.Itoa(page+1)
		}
		if page > 0 {
			data.Prev = "/search/?q=" + req + "&p=" + strconv.Itoa(page-1)
		}
		loadTemplate(w, "search", data)
	}
}
