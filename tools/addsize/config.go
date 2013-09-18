package main

const (
	PORT = "8080"

	DB_IP               = "127.0.0.1"
	DB_NAME             = "trantor"
	META_COLL           = "meta"
	BOOKS_COLL          = "books"
	TAGS_COLL           = "tags"
	VISITED_COLL        = "visited"
	DOWNLOADED_COLL     = "downloaded"
	HOURLY_VISITS_COLL  = "visits.hourly"
	DAILY_VISITS_COLL   = "visits.daily"
	MONTHLY_VISITS_COLL = "visits.monthly"
	USERS_COLL          = "users"
	NEWS_COLL           = "news"
	STATS_COLL          = "statistics"
	FS_BOOKS            = "fs_books"
	FS_IMGS             = "fs_imgs"

	PASS_SALT                 = "ImperialLibSalt"
	MINUTES_UPDATE_TAGS       = 11
	MINUTES_UPDATE_VISITED    = 41
	MINUTES_UPDATE_DOWNLOADED = 47
	MINUTES_UPDATE_HOURLY     = 31
	MINUTES_UPDATE_DAILY      = 60*12 + 7
	MINUTES_UPDATE_MONTHLY    = 60*24 + 11
	TAGS_DISPLAY              = 50
	SEARCH_ITEMS_PAGE         = 20
	NEW_ITEMS_PAGE            = 50
	NUM_NEWS                  = 10
	DAYS_NEWS_INDEXPAGE       = 15

	TEMPLATE_PATH = "templates/"
	CSS_PATH      = "css/"
	JS_PATH       = "js/"
	IMG_PATH      = "img/"

	IMG_WIDTH_BIG   = 300
	IMG_WIDTH_SMALL = 60
	IMG_QUALITY     = 80

	CHAN_SIZE = 100
)
