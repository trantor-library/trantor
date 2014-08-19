package database

import log "github.com/cihub/seelog"

import (
	"bytes"
	"code.google.com/p/go.crypto/scrypt"
	"crypto/rand"
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	Salt []byte
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
	user.Pass, user.Salt, err = hashPass(pass)
	if err != nil {
		log.Error("Error hashing password: ", err)
		return errors.New("An error happen storing the password")
	}
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
	return validatePass(pass, u.user)
}

func (u User) Role() string {
	return u.user.Role
}

func (u *User) SetPassword(pass string) error {
	if u.err != nil {
		return u.err
	}
	hash, salt, err := hashPass(pass)
	if err != nil {
		return err
	}
	return u.coll.Update(bson.M{"user": u.user.User}, bson.M{"$set": bson.M{"pass": hash, "salt": salt}})
}

func hashPass(pass string) (hash []byte, salt []byte, err error) {
	salt, err = genSalt()
	if err != nil {
		return
	}
	hash, err = calculateHash(pass, salt)
	return
}

func genSalt() ([]byte, error) {
	const (
		saltLen = 64
	)

	b := make([]byte, saltLen)
	_, err := rand.Read(b)
	return b, err
}

func validatePass(pass string, user db_user) bool {
	hash, err := calculateHash(pass, user.Salt)
	if err != nil {
		return false
	}
	return bytes.Compare(user.Pass, hash) == 0
}

func calculateHash(pass string, salt []byte) ([]byte, error) {
	const (
		N      = 16384
		r      = 8
		p      = 1
		keyLen = 32
	)

	bpass := []byte(pass)
	return scrypt.Key(bpass, salt, N, r, p, keyLen)
}
