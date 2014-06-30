package database

import log "github.com/cihub/seelog"

import (
	"bytes"
	"crypto/md5"
	"errors"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	user_coll = "users"
	pass_salt = "ImperialLibSalt"
)

type User struct {
	user db_user
	err  error
	coll *mgo.Collection
}

type db_user struct {
	User string
	Pass []byte
	Role string
}

func getUser(coll *mgo.Collection, name string) *User {
	u := new(User)
	if !validUserName(name) {
		u.err = errors.New("Invalid username")
		return u
	}

	u.coll = coll
	err := u.coll.Find(bson.M{"user": name}).One(&u.user)
	if err != nil {
		log.Warn("Error on database checking user ", name, ": ", err)
		u.err = errors.New("User not found")
		return u
	}
	return u
}

func addUser(coll *mgo.Collection, name string, pass string) error {
	if !validUserName(name) {
		return errors.New("Invalid user name")
	}
	num, err := coll.Find(bson.M{"user": name}).Count()
	if err != nil {
		log.Error("Error on database checking user ", name, ": ", err)
		return errors.New("An error happen on the database")
	}
	if num != 0 {
		return errors.New("User name already exist")
	}

	var user db_user
	user.Pass = md5Pass(pass)
	user.User = name
	user.Role = ""
	return coll.Insert(user)
}

func validUserName(name string) bool {
	return name != ""
}

func (u User) Valid(pass string) bool {
	if u.err != nil {
		return false
	}
	hash := md5Pass(pass)
	return bytes.Compare(u.user.Pass, hash) == 0
}

func (u User) Role() string {
	return u.user.Role
}

func (u *User) SetPassword(pass string) error {
	if u.err != nil {
		return u.err
	}
	hash := md5Pass(pass)
	return u.coll.Update(bson.M{"user": u.user.User}, bson.M{"$set": bson.M{"pass": hash}})
}

// FIXME: use a proper salting algorithm
func md5Pass(pass string) []byte {
	h := md5.New()
	hash := h.Sum(([]byte)(pass_salt + pass))
	return hash
}
