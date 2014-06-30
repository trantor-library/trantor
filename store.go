package main

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"git.gitorious.org/trantor/trantor.git/database"
	"gopkgs.com/unidecode.v1"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"regexp"
	"strings"
	"unicode"
)

func OpenBook(id bson.ObjectId, db *database.DB) (*epubgo.Epub, error) {
	fs := db.GetFS(FS_BOOKS)
	f, err := fs.OpenId(id)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buff, err := ioutil.ReadAll(f)
	reader := bytes.NewReader(buff)

	return epubgo.Load(reader, int64(len(buff)))
}

func StoreNewFile(name string, file io.Reader, db *database.DB) (bson.ObjectId, int64, error) {
	fs := db.GetFS(FS_BOOKS)
	fw, err := fs.Create(name)
	if err != nil {
		return "", 0, err
	}
	defer fw.Close()

	size, err := io.Copy(fw, file)
	id, _ := fw.Id().(bson.ObjectId)
	return id, size, err
}

func DeleteFile(id bson.ObjectId, db *database.DB) error {
	fs := db.GetFS(FS_BOOKS)
	return fs.RemoveId(id)
}

func DeleteCover(id bson.ObjectId, db *database.DB) error {
	fs := db.GetFS(FS_IMGS)
	return fs.RemoveId(id)
}

func DeleteBook(book database.Book, db *database.DB) {
	if book.Cover != "" {
		DeleteCover(book.Cover, db)
	}
	if book.CoverSmall != "" {
		DeleteCover(book.CoverSmall, db)
	}
	DeleteFile(book.File, db)
}

func cleanStr(str string) string {
	str = strings.Replace(str, "&#39;", "'", -1)
	exp, _ := regexp.Compile("&[^;]*;")
	str = exp.ReplaceAllString(str, "")
	exp, _ = regexp.Compile("[ ,]*$")
	str = exp.ReplaceAllString(str, "")
	return str
}

func parseAuthr(creator []string) []string {
	exp1, _ := regexp.Compile("^(.*\\( *([^\\)]*) *\\))*$")
	exp2, _ := regexp.Compile("^[^:]*: *(.*)$")
	res := make([]string, len(creator))
	for i, s := range creator {
		auth := exp1.FindStringSubmatch(s)
		if auth != nil {
			res[i] = cleanStr(strings.Join(auth[2:], ", "))
		} else {
			auth := exp2.FindStringSubmatch(s)
			if auth != nil {
				res[i] = cleanStr(auth[1])
			} else {
				res[i] = cleanStr(s)
			}
		}
	}
	return res
}

func parseDescription(description []string) string {
	str := cleanStr(strings.Join(description, "\n"))
	str = strings.Replace(str, "</p>", "\n", -1)
	exp, _ := regexp.Compile("<[^>]*>")
	str = exp.ReplaceAllString(str, "")
	str = strings.Replace(str, "&amp;", "&", -1)
	str = strings.Replace(str, "&lt;", "<", -1)
	str = strings.Replace(str, "&gt;", ">", -1)
	str = strings.Replace(str, "\\n", "\n", -1)
	return str
}

func parseSubject(subject []string) []string {
	var res []string
	for _, s := range subject {
		res = append(res, strings.Split(s, " / ")...)
	}
	return res
}

func parseDate(date []string) string {
	if len(date) == 0 {
		return ""
	}
	return strings.Replace(date[0], "Unspecified: ", "", -1)
}

func keywords(b map[string]interface{}) (k []string) {
	title, _ := b["title"].(string)
	k = tokens(title)
	author, _ := b["author"].([]string)
	for _, a := range author {
		k = append(k, tokens(a)...)
	}
	publisher, _ := b["publisher"].(string)
	k = append(k, tokens(publisher)...)
	subject, _ := b["subject"].([]string)
	for _, s := range subject {
		k = append(k, tokens(s)...)
	}
	return
}

func tokens(str string) []string {
	str = unidecode.Unidecode(str)
	str = strings.ToLower(str)
	f := func(r rune) bool {
		return unicode.IsControl(r) || unicode.IsPunct(r) || unicode.IsSpace(r)
	}
	return strings.FieldsFunc(str, f)
}
