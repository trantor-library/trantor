package main

import log "github.com/cihub/seelog"

import (
	"git.gitorious.org/trantor/trantor.git/database"
	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type handler struct {
	w    http.ResponseWriter
	r    *http.Request
	sess *Session
	db   *database.DB
}

func InitStats(database *database.DB) {
	statsChannel = make(chan statsRequest, CHAN_SIZE)
	go statsWorker(database)
}

func GatherStats(function func(handler), database *database.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Info("Query ", r.Method, " ", r.RequestURI)

		var h handler
		h.db = database.Copy()
		defer h.db.Close()

		h.w = w
		h.r = r
		h.sess = GetSession(r, h.db)
		function(h)

		statsChannel <- statsRequest{bson.Now(), mux.Vars(r), h.sess, r}
	}
}

var statsChannel chan statsRequest

type statsRequest struct {
	date time.Time
	vars map[string]string
	sess *Session
	r    *http.Request
}

func statsWorker(database *database.DB) {
	db := database.Copy()
	defer db.Close()

	for req := range statsChannel {
		stats := make(map[string]interface{})
		appendFiles(req.r, stats)
		appendMuxVars(req.vars, stats)
		appendUrl(req.r, stats)
		appendSession(req.sess, stats)
		stats["method"] = req.r.Method
		stats["date"] = req.date
		db.AddStats(stats)
	}
}

func statsHandler(h handler) {
	var data statsData
	data.S = GetStatus(h)
	data.S.Stats = true
	data.HVisits = getVisits(hourlyLabel, h.db, database.Hourly_visits)
	data.DVisits = getVisits(dailyLabel, h.db, database.Daily_visits)
	data.MVisits = getVisits(monthlyLabel, h.db, database.Monthly_visits)
	data.HDownloads = getVisits(hourlyLabel, h.db, database.Hourly_downloads)
	data.DDownloads = getVisits(dailyLabel, h.db, database.Daily_downloads)
	data.MDownloads = getVisits(monthlyLabel, h.db, database.Monthly_downloads)

	loadTemplate(h.w, "stats", data)
}

type statsData struct {
	S          Status
	HVisits    []visitData
	DVisits    []visitData
	MVisits    []visitData
	HDownloads []visitData
	DDownloads []visitData
	MDownloads []visitData
}

type visitData struct {
	Label string
	Count int
}

func hourlyLabel(date time.Time) string {
	return strconv.Itoa(date.Hour() + 1)
}

func dailyLabel(date time.Time) string {
	return strconv.Itoa(date.Day())
}

func monthlyLabel(date time.Time) string {
	return date.Month().String()
}

func getVisits(funcLabel func(time.Time) string, db *database.DB, visitType database.VisitType) []visitData {
	var visits []visitData

	visit, err := db.GetVisits(visitType)
	if err != nil {
		log.Warn("GetVisits error (", visitType, "): ", err)
	}
	for _, v := range visit {
		var elem visitData
		elem.Label = funcLabel(v.Date.UTC())
		elem.Count = v.Count
		visits = append(visits, elem)
	}

	return visits
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
