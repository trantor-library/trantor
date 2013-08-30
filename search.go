package main

import (
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
)

func buildQuery(q string) bson.M {
	var reg []bson.RegEx
	query := bson.M{"active": true}
	words := strings.Split(q, " ")
	for _, w := range words {
		tag := strings.SplitN(w, ":", 2)
		if len(tag) > 1 {
			query[tag[0]] = bson.RegEx{tag[1], "i"}
		} else {
			reg = append(reg, bson.RegEx{w, "i"})
		}
	}
	if len(reg) > 0 {
		query["keywords"] = bson.M{"$all": reg}
	}
	return query
}

type searchData struct {
	S     Status
	Found int
	Books []Book
	Page  int
	Next  string
	Prev  string
}

func searchHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := strings.Join(r.Form["q"], " ")
	page := 0
	if len(r.Form["p"]) != 0 {
		page, err = strconv.Atoi(r.Form["p"][0])
		if err != nil {
			page = 0
		}
	}
	res, num, _ := db.GetBooks(buildQuery(req), SEARCH_ITEMS_PAGE, page*SEARCH_ITEMS_PAGE)

	var data searchData
	data.S = GetStatus(w, r)
	data.S.Search = req
	data.Books = res
	data.Found = num
	data.Page = page + 1
	if num > (page+1)*SEARCH_ITEMS_PAGE {
		data.Next = "/search/?q=" + req + "&p=" + strconv.Itoa(page+1)
	}
	if page > 0 {
		data.Prev = "/search/?q=" + req + "&p=" + strconv.Itoa(page-1)
	}

	format := r.Form["fmt"]
	if (len(format) > 0) && (format[0] == "rss") {
		loadTxtTemplate(w, "search_rss.xml", data)
	} else {
		loadTemplate(w, "search", data)
	}
}
