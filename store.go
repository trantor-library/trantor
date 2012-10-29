package main

import (
	"git.gitorious.org/go-pkg/epub.git"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

func ParseFile(path string) (string, error) {
	book := map[string]interface{}{}

	e, err := epub.Open(NEW_PATH+path, 0)
	if err != nil {
		return "", err
	}
	defer e.Close()

	title := cleanStr(strings.Join(e.Metadata(epub.EPUB_TITLE), ", "))
	book["title"] = title
	book["author"] = parseAuthr(e.Metadata(epub.EPUB_CREATOR))
	book["contributor"] = cleanStr(strings.Join(e.Metadata(epub.EPUB_CONTRIB), ", "))
	book["publisher"] = cleanStr(strings.Join(e.Metadata(epub.EPUB_PUBLISHER), ", "))
	book["description"] = parseDescription(e.Metadata(epub.EPUB_DESCRIPTION))
	book["subject"] = parseSubject(e.Metadata(epub.EPUB_SUBJECT))
	book["date"] = parseDate(e.Metadata(epub.EPUB_DATE))
	book["lang"] = e.Metadata(epub.EPUB_LANG)
	book["type"] = strings.Join(e.Metadata(epub.EPUB_TYPE), ", ")
	book["format"] = strings.Join(e.Metadata(epub.EPUB_FORMAT), ", ")
	book["source"] = strings.Join(e.Metadata(epub.EPUB_SOURCE), ", ")
	book["relation"] = strings.Join(e.Metadata(epub.EPUB_RELATION), ", ")
	book["coverage"] = strings.Join(e.Metadata(epub.EPUB_COVERAGE), ", ")
	book["rights"] = strings.Join(e.Metadata(epub.EPUB_RIGHTS), ", ")
	book["meta"] = strings.Join(e.Metadata(epub.EPUB_META), ", ")
	book["path"] = path
	cover, coverSmall := getCover(e, title)
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

	const size = 1024
	var n int = size
	buff := make([]byte, size)
	for n == size {
		n, err = file.Read(buff)
		fw.Write(buff)
	}
	return path, nil
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

func storeImg(img []byte, title, extension string) (string, string) {
	r, _ := utf8.DecodeRuneInString(title)
	folder := string(r)
	if _, err := os.Stat(COVER_PATH + folder); err != nil {
		err = os.Mkdir(COVER_PATH+folder, os.ModePerm)
		if err != nil {
			return "", ""
		}
	}
	imgPath := validFileName(COVER_PATH, title, extension)

	/* store img on disk */
	file, err := os.Create(COVER_PATH + imgPath)
	if err != nil {
		return "", ""
	}
	defer file.Close()
	file.Write(img)

	/* resize img */
	resize := append(strings.Split(RESIZE_CMD, " "), COVER_PATH+imgPath, COVER_PATH+imgPath)
	cmd := exec.Command(resize[0], resize[1:]...)
	cmd.Run()
	imgPathSmall := validFileName(COVER_PATH, title, "_small"+extension)
	resize = append(strings.Split(RESIZE_THUMB_CMD, " "), COVER_PATH+imgPath, COVER_PATH+imgPathSmall)
	cmd = exec.Command(resize[0], resize[1:]...)
	cmd.Run()
	return imgPath, imgPathSmall
}

func getCover(e *epub.Epub, title string) (string, string) {
	/* Try first common names */
	for _, p := range []string{"cover.jpg", "Images/cover.jpg", "cover.jpeg", "cover1.jpg", "cover1.jpeg"} {
		img := e.Data(p)
		if len(img) != 0 {
			return storeImg(img, title, ".jpg")
		}
	}

	/* search for img on the text */
	exp, _ := regexp.Compile("<ima?g.*[(src)(href)]=[\"']([^\"']*(\\.[^\\.\"']*))[\"']")
	it := e.Iterator(epub.EITERATOR_SPINE)
	defer it.Close()
	var err error = nil
	txt := it.Curr()
	for err == nil {
		res := exp.FindStringSubmatch(txt)
		if res != nil {
			urlPart := strings.Split(it.CurrUrl(), "/")
			url := strings.Join(urlPart[:len(urlPart)-1], "/")
			if res[1][:3] == "../" {
				res[1] = res[1][3:]
				url = strings.Join(urlPart[:len(urlPart)-2], "/")
			}
			res[1] = strings.Replace(res[1], "%20", " ", -1)
			res[1] = strings.Replace(res[1], "%27", "'", -1)
			res[1] = strings.Replace(res[1], "%28", "(", -1)
			res[1] = strings.Replace(res[1], "%29", ")", -1)
			if url == "" {
				url = res[1]
			} else {
				url = url + "/" + res[1]
			}

			img := e.Data(url)
			if len(img) != 0 {
				return storeImg(img, title, res[2])
			}
		}
		txt, err = it.Next()
	}
	return "", ""
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
