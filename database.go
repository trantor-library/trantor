package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sort"
)

type Book struct {
	Id          string `bson:"_id"`
	Title       string
	Author      []string
	Contributor string
	Publisher   string
	Description string
	Subject     []string
	Date        string
	Lang        []string
	Type        string
	Format      string
	Source      string
	Relation    string
	Coverage    string
	Rights      string
	Meta        string
	Path        string
	Cover       string
	CoverSmall  string
	Keywords    []string
}

/* optional parameters: length and start index
 * 
 * Returns: list of books, number found and err
 */
func GetBook(coll *mgo.Collection, query bson.M, r ...int) (books []Book, num int, err error) {
	var start, length int
	if len(r) > 0 {
		length = r[0]
		if len(r) > 1 {
			start = r[1]
		}
	}
	q := coll.Find(query).Sort("-_id")
	num, err = q.Count()
	if err != nil {
		return
	}
	if start != 0 {
		q = q.Skip(start)
	}
	if length != 0 {
		q = q.Limit(length)
	}

	err = q.All(&books)
	for i, b := range books {
		books[i].Id = bson.ObjectId(b.Id).Hex()
	}
	return

}

type tagsList []struct {
	Subject string "_id"
	Count   int    "value"
}

func (t tagsList) Len() int {
	return len(t)
}
func (t tagsList) Less(i, j int) bool {
	return t[i].Count > t[j].Count
}
func (t tagsList) Swap(i, j int) {
	aux := t[i]
	t[i] = t[j]
	t[j] = aux
}

func GetTags(coll *mgo.Collection) (tagsList, error) {
	// TODO: cache the tags
	var mr mgo.MapReduce
	mr.Map = "function() { " +
		"this.subject.forEach(function(s) { emit(s, 1); });" +
		"}"
	mr.Reduce = "function(tag, vals) { " +
		"var count = 0;" +
		"vals.forEach(function() { count += 1; });" +
		"return count;" +
		"}"
	var result tagsList
	_, err := coll.Find(nil).MapReduce(&mr, &result)
	if err == nil {
		sort.Sort(result)
	}
	return result, err
}
