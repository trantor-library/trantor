package main

import (
	"code.google.com/p/gopass"
	"crypto/md5"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
)

const (
	IP         = "127.0.0.1"
	DB_NAME    = "trantor"
	USERS_COLL = "users"
	PASS_SALT  = "ImperialLibSalt"
)

func main() {
	session, err := mgo.Dial(IP)
	if err != nil {
		panic(err)
	}
	defer session.Close()
	coll := session.DB(DB_NAME).C(USERS_COLL)

	user := os.Args[1]
	pass, err := gopass.GetPass("Password: ")
	if err != nil {
		panic(err)
	}
	h := md5.New()
	hash := h.Sum(([]byte)(PASS_SALT + pass))
	fmt.Println(user, " - ", hash)
	err = coll.Insert(bson.M{"user": user, "pass": hash, "role": "admin"})
	if err != nil {
		panic(err)
	}
}
