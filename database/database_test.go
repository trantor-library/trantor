package database

import "testing"

const (
	test_coll = "test_trantor"
	test_host = "127.0.0.1"
)

func TestInit(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.Close()
}

func TestCopy(t *testing.T) {
	db := Init(test_host, test_coll)
	defer db.del()

	db2 := db.Copy()

	if db.name != db2.name {
		t.Errorf("Names don't match")
	}
	names1, err := db.session.DatabaseNames()
	if err != nil {
		t.Errorf("Error on db1: ", err)
	}
	names2, err := db2.session.DatabaseNames()
	if err != nil {
		t.Errorf("Error on db1: ", err)
	}
	if len(names1) != len(names2) {
		t.Errorf("len(names) don't match")
	}
	for i, _ := range names1 {
		if names1[i] != names2[i] {
			t.Errorf("Names don't match")
		}
	}
}
