package main

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"regexp"
	"strings"
)

func OpenBook(id bson.ObjectId) (*epubgo.Epub, error) {
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

func StoreNewFile(name string, file io.Reader) (bson.ObjectId, error) {
	fs := db.GetFS(FS_BOOKS)
	fw, err := fs.Create(name)
	if err != nil {
		return "", err
	}
	defer fw.Close()

	_, err = io.Copy(fw, file)
	id, _ := fw.Id().(bson.ObjectId)
	return id, err
}

func DeleteFile(id bson.ObjectId) error {
	fs := db.GetFS(FS_BOOKS)
	return fs.RemoveId(id)
}

func DeleteCover(id bson.ObjectId) error {
	fs := db.GetFS(FS_IMGS)
	return fs.RemoveId(id)
}

func DeleteBook(book Book) {
	if book.Cover != "" {
		DeleteCover(book.Cover)
	}
	if book.CoverSmall != "" {
		DeleteCover(book.CoverSmall)
	}
	DeleteFile(book.File)
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
	k = strings.Split(title, " ")
	author, _ := b["author"].([]string)
	for _, a := range author {
		k = append(k, strings.Split(a, " ")...)
	}
	publisher, _ := b["publisher"].(string)
	k = append(k, strings.Split(publisher, " ")...)
	subject, _ := b["subject"].([]string)
	k = append(k, subject...)
	return
}
