package main

import (
	"code.google.com/p/gopass"
	"git.gitorious.org/trantor/trantor.git/database"
	"os"
)

const (
	DB_IP   = "127.0.0.1"
	DB_NAME = "trantor"
)

func main() {
	db := database.Init(DB_IP, DB_NAME)
	defer db.Close()

	user := os.Args[1]
	pass, err := gopass.GetPass("Password: ")
	if err != nil {
		panic(err)
	}

	err = db.AddUser(user, pass)
	if err != nil {
		panic(err)
	}
}
