package database

import "testing"

const (
	name, pass = "user", "mypass"
)

func TestUserEmpty(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	if db.User("").Valid("") {
		t.Errorf("user.Valid() with an empty password return true")
	}
}

func TestAddUser(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddUser(t, db)
	if !db.User(name).Valid(pass) {
		t.Errorf("user.Valid() return false for a valid user")
	}
}

func TestEmptyUsername(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	tAddUser(t, db)
	if db.User("").Valid(pass) {
		t.Errorf("user.Valid() return true for an invalid user")
	}
}

func tAddUser(t *testing.T, db *DB) {
	err := db.AddUser(name, pass)
	if err != nil {
		t.Errorf("db.Adduser(", name, ", ", pass, ") return an error: ", err)
	}
}
