package main

const (
	PORT              = "8080"
	DB_IP             = "127.0.0.1"
	DB_NAME           = "trantor"
	BOOKS_COLL        = "books"
	USERS_COLL        = "users"
	PASS_SALT         = "ImperialLibSalt"
	TAGS_DISPLAY      = 50
	SEARCH_ITEMS_PAGE = 20
	NEW_ITEMS_PAGE    = 50
	TEMPLATE_PATH     = "templates/"
	BOOKS_PATH        = "books/"
	COVER_PATH        = "cover/"
	NEW_PATH          = "new/"
	CSS_PATH          = "css/"
	JS_PATH           = "js/"
	IMG_PATH          = "img/"
	RESIZE_CMD        = "/usr/bin/convert -resize 300 -quality 60 "
	RESIZE_THUMB_CMD  = "/usr/bin/convert -resize 60 -quality 60 "
)
