package main

const (
	PORT = "8080"

	DB_IP     = "127.0.0.1"
	DB_NAME   = "trantor"
	META_COLL = "meta"

	EPUB_FILE        = "book.epub"
	COVER_FILE       = "cover.jpg"
	COVER_SMALL_FILE = "coverSmall.jpg"

	MINUTES_UPDATE_TAGS       = 11
	MINUTES_UPDATE_VISITED    = 41
	MINUTES_UPDATE_DOWNLOADED = 47
	MINUTES_UPDATE_HOURLY_V   = 31
	MINUTES_UPDATE_DAILY_V    = 60*12 + 7
	MINUTES_UPDATE_MONTHLY_V  = 60*24 + 11
	MINUTES_UPDATE_HOURLY_D   = 29
	MINUTES_UPDATE_DAILY_D    = 60*12 + 13
	MINUTES_UPDATE_MONTHLY_D  = 60*24 + 17
	MINUTES_UPDATE_LOGGER     = 5
	BOOKS_FRONT_PAGE          = 6
	SEARCH_ITEMS_PAGE         = 20
	NEW_ITEMS_PAGE            = 50
	NUM_NEWS                  = 10
	DAYS_NEWS_INDEXPAGE       = 15

	STORE_PATH       = "store/"
	TEMPLATE_PATH    = "templates/"
	CSS_PATH         = "css/"
	JS_PATH          = "js/"
	IMG_PATH         = "img/"
	ROBOTS_PATH      = "robots.txt"
	DESCRIPTION_PATH = "description.json"
	LOGGER_CONFIG    = "logger.xml"

	IMG_WIDTH_BIG   = 300
	IMG_WIDTH_SMALL = 60
	IMG_QUALITY     = 80

	CHAN_SIZE = 100
)
