package main

import (
	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strings"
	"time"
)

func InitStats() {
	statsChannel = make(chan statsRequest, CHAN_SIZE)
	go statsWorker()
}

func GatherStats(function func(http.ResponseWriter, *http.Request, *Session)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess := GetSession(r)
		function(w, r, sess)
		statsChannel <- statsRequest{bson.Now(), mux.Vars(r), sess, r}
	}
}

var statsChannel chan statsRequest

type statsRequest struct {
	date time.Time
	vars map[string]string
	sess *Session
	r    *http.Request
}

func statsWorker() {
	for req := range statsChannel {
		stats := make(map[string]interface{})
		appendFiles(req.r, stats)
		appendMuxVars(req.vars, stats)
		appendUrl(req.r, stats)
		appendSession(req.sess, stats)
		stats["method"] = req.r.Method
		stats["date"] = req.date
		db.InsertStats(stats)
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data statsData
	data.S = GetStatus(w, r)
	data.Visits, _ = db.GetMonthVisits()
	loadTemplate(w, "stats", data)
}

type statsData struct {
	S      Status
	Visits []visits
}

func appendFiles(r *http.Request, stats map[string]interface{}) {
	if r.Method == "POST" && r.MultipartForm != nil {
		files := r.MultipartForm.File
		for key := range files {
			list := make([]string, len(files[key]))
			for i, f := range files[key] {
				list[i] = f.Filename
			}
			stats[key] = list
		}
	}
}

func appendMuxVars(vars map[string]string, stats map[string]interface{}) {
	for key, value := range vars {
		switch {
		case key == "id":
			if bson.IsObjectIdHex(value) {
				stats["id"] = bson.ObjectIdHex(value)
			}
		case key == "ids":
			var objectIds []bson.ObjectId
			ids := strings.Split(value, "/")
			for _, id := range ids {
				if bson.IsObjectIdHex(value) {
					objectIds = append(objectIds, bson.ObjectIdHex(id))
				}
			}
			if len(objectIds) > 0 {
				stats["ids"] = objectIds
				stats["id"] = objectIds[0]
			}
		default:
			stats[key] = value
		}
	}
}

func appendUrl(r *http.Request, stats map[string]interface{}) {
	for key, value := range r.URL.Query() {
		stats[key] = value
	}
	stats["host"] = r.Host
	stats["path"] = r.URL.Path
	pattern := strings.Split(r.URL.Path, "/")
	if len(pattern) > 1 && pattern[1] != "" {
		stats["section"] = pattern[1]
	} else {
		stats["section"] = "/"
	}
}

func appendSession(sess *Session, stats map[string]interface{}) {
	stats["session"] = sess.Id()
	if sess.User != "" {
		stats["user"] = sess.User
	}
}
