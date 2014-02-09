package main

import (
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
	db   *DB
}

func InitStats(database *DB) {
	statsChannel = make(chan statsRequest, CHAN_SIZE)
	go statsWorker(database)
}

func GatherStats(function func(handler), database *DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
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

func statsWorker(database *DB) {
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
		db.InsertStats(stats)
	}
}

func statsHandler(h handler) {
	var data statsData
	data.S = GetStatus(h)
	data.S.Stats = true
	data.HVisits = getHourlyVisits(h.db)
	data.DVisits = getDailyVisits(h.db)
	data.MVisits = getMonthlyVisits(h.db)
	data.HDownloads = getHourlyDownloads(h.db)
	data.DDownloads = getDailyDownloads(h.db)
	data.MDownloads = getMonthlyDownloads(h.db)

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

func getHourlyVisits(db *DB) []visitData {
	var visits []visitData

	visit, _ := db.GetHourVisits()
	for _, v := range visit {
		var elem visitData
		hour := time.Unix(v.Date/1000, 0).UTC().Hour()
		elem.Label = strconv.Itoa(hour + 1)
		elem.Count = v.Count
		visits = append(visits, elem)
	}

	return visits
}

func getDailyVisits(db *DB) []visitData {
	var visits []visitData

	visit, _ := db.GetDayVisits()
	for _, v := range visit {
		var elem visitData
		day := time.Unix(v.Date/1000, 0).UTC().Day()
		elem.Label = strconv.Itoa(day)
		elem.Count = v.Count
		visits = append(visits, elem)
	}

	return visits
}

func getMonthlyVisits(db *DB) []visitData {
	var visits []visitData

	visit, _ := db.GetMonthVisits()
	for _, v := range visit {
		var elem visitData
		month := time.Unix(v.Date/1000, 0).UTC().Month()
		elem.Label = month.String()
		elem.Count = v.Count
		visits = append(visits, elem)
	}

	return visits
}

func getHourlyDownloads(db *DB) []visitData {
	var visits []visitData

	visit, _ := db.GetHourDownloads()
	for _, v := range visit {
		var elem visitData
		hour := time.Unix(v.Date/1000, 0).UTC().Hour()
		elem.Label = strconv.Itoa(hour + 1)
		elem.Count = v.Count
		visits = append(visits, elem)
	}

	return visits
}

func getDailyDownloads(db *DB) []visitData {
	var visits []visitData

	visit, _ := db.GetDayDownloads()
	for _, v := range visit {
		var elem visitData
		day := time.Unix(v.Date/1000, 0).UTC().Day()
		elem.Label = strconv.Itoa(day)
		elem.Count = v.Count
		visits = append(visits, elem)
	}

	return visits
}

func getMonthlyDownloads(db *DB) []visitData {
	var visits []visitData

	visit, _ := db.GetMonthDownloads()
	for _, v := range visit {
		var elem visitData
		month := time.Unix(v.Date/1000, 0).UTC().Month()
		elem.Label = month.String()
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
