package main

import (
	"time"
)

func InitTasks(db *DB) {
	periodicTask(db.UpdateTags, MINUTES_UPDATE_TAGS*time.Minute)
	periodicTask(db.UpdateMostVisited, MINUTES_UPDATE_VISITED*time.Minute)
	periodicTask(db.UpdateDownloadedBooks, MINUTES_UPDATE_DOWNLOADED*time.Minute)
	periodicTask(db.UpdateHourVisits, MINUTES_UPDATE_HOURLY*time.Minute)
	periodicTask(db.UpdateDayVisits, MINUTES_UPDATE_DAILY*time.Minute)
	periodicTask(db.UpdateMonthVisits, MINUTES_UPDATE_MONTHLY*time.Minute)
	periodicTask(db.UpdateHourDownloads, MINUTES_UPDATE_HOURLY*time.Minute)
	periodicTask(db.UpdateDayDownloads, MINUTES_UPDATE_DAILY*time.Minute)
	periodicTask(db.UpdateMonthDownloads, MINUTES_UPDATE_MONTHLY*time.Minute)
}

func periodicTask(task func() error, periodicity time.Duration) {
	go tasker(task, periodicity)
}

func tasker(task func() error, periodicity time.Duration) {
	for true {
		time.Sleep(periodicity)
		task()
	}
}
