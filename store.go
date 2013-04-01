package main

import (
	"git.gitorious.org/go-pkg/epubgo.git"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ParseFile(path string) (string, error) {
	book := map[string]interface{}{}

	e, err := epubgo.Open(NEW_PATH + path)
	if err != nil {
		return "", err
	}
	defer e.Close()

	for _, m := range e.MetadataFields() {
		data, err := e.Metadata(m)
		if err != nil {
			continue
		}
		switch m {
		case "creator":
			book["author"] = parseAuthr(data)
		case "description":
			book[m] = parseDescription(data)
		case "subject":
			book[m] = parseSubject(data)
		case "date":
			book[m] = parseDate(data)
		case "language":
			book["lang"] = data
		case "title", "contributor", "publisher":
			book[m] = cleanStr(strings.Join(data, ", "))
		default:
			book[m] = strings.Join(data, ", ")
		}
	}
	title, _ := book["title"].(string)
	book["path"] = path
	cover, coverSmall := GetCover(e, title)
	book["cover"] = cover
	book["coversmall"] = coverSmall
	book["keywords"] = keywords(book)

	db.InsertBook(book)
	return title, nil
}

func StoreNewFile(name string, file io.Reader) (string, error) {
	path := storePath(name)
	fw, err := os.Create(NEW_PATH + path)
	if err != nil {
		return "", err
	}
	defer fw.Close()

	_, err = io.Copy(fw, file)
	return path, err
}

func StoreBook(book Book) (path string, err error) {
	title := book.Title
	path = validFileName(BOOKS_PATH, title, ".epub")

	oldPath := NEW_PATH + book.Path
	r, _ := utf8.DecodeRuneInString(title)
	folder := string(r)
	if _, err = os.Stat(BOOKS_PATH + folder); err != nil {
		err = os.Mkdir(BOOKS_PATH+folder, os.ModePerm)
		if err != nil {
			log.Println("Error creating", BOOKS_PATH+folder, ":", err.Error())
			return
		}
	}
	cmd := exec.Command("mv", oldPath, BOOKS_PATH+path)
	err = cmd.Run()
	return
}

func DeleteBook(book Book) {
	if book.Cover != "" {
		os.RemoveAll(book.Cover[1:])
	}
	if book.CoverSmall != "" {
		os.RemoveAll(book.CoverSmall[1:])
	}
	os.RemoveAll(book.Path)
}

func validFileName(path string, title string, extension string) string {
	title = strings.Replace(title, "/", "_", -1)
	title = strings.Replace(title, "?", "_", -1)
	title = strings.Replace(title, "#", "_", -1)
	r, _ := utf8.DecodeRuneInString(title)
	folder := string(r)
	file := folder + "/" + title + extension
	_, err := os.Stat(path + file)
	for i := 0; err == nil; i++ {
		file = folder + "/" + title + "_" + strconv.Itoa(i) + extension
		_, err = os.Stat(path + file)
	}
	return file
}

func storePath(name string) string {
	path := name
	_, err := os.Stat(NEW_PATH + path)
	for i := 0; err == nil; i++ {
		path = strconv.Itoa(i) + "_" + name
		_, err = os.Stat(NEW_PATH + path)
	}
	return path
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
	str := cleanStr(strings.Join(description, ", "))
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
