package main

import (
	"time"
)

func InitTasks(db *DB) {
	initTagsTask(db)
	initVisitedTask(db)
	initDownloadedTask(db)
	initHourVisitsTask(db)
	initDayVisitsTask(db)
	initMonthVisitsTask(db)
	initHourDownloadsTask(db)
	initDayDownloadsTask(db)
	initMonthDownloadsTask(db)
}

func initTagsTask(db *DB) {
	updateTags := func() {
		db.UpdateTags()
	}
	periodicTask(updateTags, MINUTES_UPDATE_TAGS*time.Minute)
}

func initVisitedTask(db *DB) {
	updateVisited := func() {
		db.UpdateMostVisited()
	}
	periodicTask(updateVisited, MINUTES_UPDATE_VISITED*time.Minute)
}

func initDownloadedTask(db *DB) {
	updateDownloaded := func() {
		db.UpdateDownloadedBooks()
	}
	periodicTask(updateDownloaded, MINUTES_UPDATE_DOWNLOADED*time.Minute)
}

func initHourVisitsTask(db *DB) {
	updateHourVisits := func() {
		db.UpdateHourVisits()
	}
	periodicTask(updateHourVisits, MINUTES_UPDATE_HOURLY*time.Minute)
}

func initDayVisitsTask(db *DB) {
	updateDayVisits := func() {
		db.UpdateDayVisits()
	}
	periodicTask(updateDayVisits, MINUTES_UPDATE_HOURLY*time.Minute)
}

func initMonthVisitsTask(db *DB) {
	updateMonthVisits := func() {
		db.UpdateMonthVisits()
	}
	periodicTask(updateMonthVisits, MINUTES_UPDATE_HOURLY*time.Minute)
}

func initHourDownloadsTask(db *DB) {
	updateHourDownloads := func() {
		db.UpdateHourDownloads()
	}
	periodicTask(updateHourDownloads, MINUTES_UPDATE_HOURLY*time.Minute)
}

func initDayDownloadsTask(db *DB) {
	updateDayDownloads := func() {
		db.UpdateDayDownloads()
	}
	periodicTask(updateDayDownloads, MINUTES_UPDATE_HOURLY*time.Minute)
}

func initMonthDownloadsTask(db *DB) {
	updateMonthDownloads := func() {
		db.UpdateMonthDownloads()
	}
	periodicTask(updateMonthDownloads, MINUTES_UPDATE_HOURLY*time.Minute)
}

func periodicTask(task func(), periodicity time.Duration) {
	go tasker(task, periodicity)
}

func tasker(task func(), periodicity time.Duration) {
	for true {
		time.Sleep(periodicity)
		task()
	}
}
