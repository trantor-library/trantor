package main

import (
	"fmt"
	"git.gitorious.org/go-pkg/epub.git"
	"labix.org/v2/mgo"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	IP         = "127.0.0.1"
	DB_NAME    = "trantor"
	BOOKS_COLL = "books"
	PATH       = "books/"
	NEW_PATH    = "new/"
	COVER_PATH = "cover/"
	RESIZE     = "/usr/bin/convert -resize 300 -quality 60 "
	RESIZE_THUMB = "/usr/bin/convert -resize 60 -quality 60 "
)

func getCover(e *epub.Epub, path string) (string, string) {
	exp, _ := regexp.Compile("<img.*src=[\"']([^\"']*(\\.[^\\.\"']*))[\"']")
	it := e.Iterator(epub.EITERATOR_SPINE)
	defer it.Close()

	var err error = nil
	txt := it.Curr()
	for err == nil {
		res := exp.FindStringSubmatch(txt)
		if res != nil {
			folder := COVER_PATH + path[:1] + "/"
			os.Mkdir(folder, os.ModePerm)
			imgPath := folder + path + res[2]
			f, _ := os.Create(imgPath)
			f.Write(e.Data(res[1]))
			resize := append(strings.Split(RESIZE, " "), imgPath, imgPath)
			cmd := exec.Command(resize[0], resize[1:]...)
			cmd.Run()
			imgPathSmall := folder + path + "_small" + res[2]
			resize = append(strings.Split(RESIZE_THUMB, " "), imgPath, imgPathSmall)
			cmd = exec.Command(resize[0], resize[1:]...)
			cmd.Run()
			return "/" + imgPath, "/" + imgPathSmall
		}
		txt, err = it.Next()
	}
	return "", ""
}

func parseAuthr(creator []string) []string {
	exp1, _ := regexp.Compile("^(.*\\( *([^\\)]*) *\\))*$")
	exp2, _ := regexp.Compile("^[^:]*: *(.*)$")
	var res []string //TODO: can be predicted the lenght
	for _, s := range creator {
		auth := exp1.FindStringSubmatch(s)
		if auth != nil {
			res = append(res, strings.Join(auth[2:], ", "))
		} else {
			auth := exp2.FindStringSubmatch(s)
			if auth != nil {
				res = append(res, auth[1])
			} else {
				res = append(res, s)
			}
		}
	}
	return res
}

func parseSubject(subject []string) []string {
	var res []string
	for _, s := range subject {
		res = append(res, strings.Split(s, " / ")...) // FIXME
	}
	return res
}

func keywords(b Book) (k []string) {
	k = strings.Split(b.Title, " ")
	for _, a := range b.Author {
		k = append(k, strings.Split(a, " ")...)
	}
	k = append(k, strings.Split(b.Publisher, " ")...)
	k = append(k, b.Subject...)
	return
}

func store(coll *mgo.Collection, path string) {
	var book Book

	e, err := epub.Open(NEW_PATH+path, 0)
	if err != nil {
		fmt.Println(path)
		panic(err) // TODO: do something
	}
	defer e.Close()

	book.Title = strings.Join(e.Metadata(epub.EPUB_TITLE), ", ")
	book.Author = parseAuthr(e.Metadata(epub.EPUB_CREATOR))
	book.Contributor = strings.Join(e.Metadata(epub.EPUB_CONTRIB), ", ")
	book.Publisher = strings.Join(e.Metadata(epub.EPUB_PUBLISHER), ", ")
	book.Description = strings.Join(e.Metadata(epub.EPUB_DESCRIPTION), ", ")
	book.Subject = parseSubject(e.Metadata(epub.EPUB_SUBJECT))
	book.Date = strings.Join(e.Metadata(epub.EPUB_DATE), ", ")
	book.Lang = e.Metadata(epub.EPUB_LANG)
	book.Type = strings.Join(e.Metadata(epub.EPUB_TYPE), ", ")
	book.Format = strings.Join(e.Metadata(epub.EPUB_FORMAT), ", ")
	book.Source = strings.Join(e.Metadata(epub.EPUB_SOURCE), ", ")
	book.Relation = strings.Join(e.Metadata(epub.EPUB_RELATION), ", ")
	book.Coverage = strings.Join(e.Metadata(epub.EPUB_COVERAGE), ", ")
	book.Rights = strings.Join(e.Metadata(epub.EPUB_RIGHTS), ", ")
	book.Meta = strings.Join(e.Metadata(epub.EPUB_META), ", ")
	book.Path = PATH + path[:1] + "/" + path
	book.Cover, book.CoverSmall = getCover(e, path)
	book.Keywords = keywords(book)
	coll.Insert(book)

	os.Mkdir(PATH + path[:1], os.ModePerm)
	cmd := exec.Command("mv", NEW_PATH+path, book.Path)
	cmd.Run()
}

func main() {
	session, err := mgo.Dial(IP)
	if err != nil {
		panic(err) // TODO: do something
	}
	defer session.Close()
	coll := session.DB(DB_NAME).C(BOOKS_COLL)

	f, err := os.Open(NEW_PATH)
	if err != nil {
		fmt.Println(NEW_PATH)
		panic(err) // TODO: do something
	}
	names, err := f.Readdirnames(0)
	if err != nil {
		panic(err) // TODO: do something
	}

	for _, name := range names {
		store(coll, name)
	}
}
