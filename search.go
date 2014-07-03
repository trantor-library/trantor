package main

import (
	"git.gitorious.org/trantor/trantor.git/database"
	"net/http"
	"strconv"
	"strings"
)

type searchData struct {
	S         Status
	Found     int
	Books     []database.Book
	ItemsPage int
	Page      int
	Next      string
	Prev      string
}

func searchHandler(h handler) {
	err := h.r.ParseForm()
	if err != nil {
		http.Error(h.w, err.Error(), http.StatusInternalServerError)
		return
	}
	req := strings.Join(h.r.Form["q"], " ")
	page := 0
	if len(h.r.Form["p"]) != 0 {
		page, err = strconv.Atoi(h.r.Form["p"][0])
		if err != nil {
			page = 0
		}
	}
	items_page := itemsPage(h.r)
	res, num, _ := h.db.GetBooks(req, items_page, page*items_page)

	var data searchData
	data.S = GetStatus(h)
	data.S.Search = req
	data.Books = res
	data.ItemsPage = items_page
	data.Found = num
	data.Page = page + 1
	if num > (page+1)*items_page {
		data.Next = "/search/?q=" + req + "&p=" + strconv.Itoa(page+1) + "&num=" + strconv.Itoa(items_page)
	}
	if page > 0 {
		data.Prev = "/search/?q=" + req + "&p=" + strconv.Itoa(page-1) + "&num=" + strconv.Itoa(items_page)
	}

	format := h.r.Form["fmt"]
	if (len(format) > 0) && (format[0] == "rss") {
		loadTxtTemplate(h.w, "search_rss.xml", data)
	} else {
		loadTemplate(h.w, "search", data)
	}
}

func itemsPage(r *http.Request) int {
	if len(r.Form["num"]) > 0 {
		items_page, err := strconv.Atoi(r.Form["num"][0])
		if err == nil {
			return items_page
		}
	}
	return SEARCH_ITEMS_PAGE
}
