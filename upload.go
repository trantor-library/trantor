package main

import (
	"git.gitorious.org/go-pkg/epub.git"
	"labix.org/v2/mgo"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func storePath(name string) string {
	path := NEW_PATH + name
	_, err := os.Stat(path)
	for i := 0; err == nil; i++ {
		path = NEW_PATH + strconv.Itoa(i) + "_" + name
		_, err = os.Stat(path)
	}
	return path
}

func storeFiles(r *http.Request) ([]string, error) {
	r.ParseMultipartForm(20000000)
	filesForm := r.MultipartForm.File["epub"]
	paths := make([]string, 0, len(filesForm))
	for _, f := range filesForm {
		file, err := f.Open()
		if err != nil {
			return paths, err
		}
		defer file.Close()

		path := storePath(f.Filename)
		fw, err := os.Create(path)
		if err != nil {
			return paths, err
		}
		defer fw.Close()

		const size = 1024
		var n int = size
		buff := make([]byte, size)
		for n == size {
			n, err = file.Read(buff)
			fw.Write(buff)
		}
		paths = append(paths, path)
	}
	return paths, nil
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
	name := title
	folder := COVER_PATH + name[:1] + "/"
	os.Mkdir(folder, os.ModePerm)
	imgPath := folder + name + extension
	_, err := os.Stat(imgPath)
	for i := 0; err == nil; i++ {
		name = title + "_" + strconv.Itoa(i)
		imgPath = folder + name + extension
		_, err = os.Stat(imgPath)
	}

	/* store img on disk */
	file, _ := os.Create(imgPath)
	defer file.Close()
	file.Write(img)

	/* resize img */
	resize := append(strings.Split(RESIZE_CMD, " "), imgPath, imgPath)
	cmd := exec.Command(resize[0], resize[1:]...)
	cmd.Run()
	imgPathSmall := folder + name + "_small" + extension
	resize = append(strings.Split(RESIZE_THUMB_CMD, " "), imgPath, imgPathSmall)
	cmd = exec.Command(resize[0], resize[1:]...)
	cmd.Run()
	return "/" + imgPath, "/" + imgPathSmall
}

func getCover(e *epub.Epub, title string) (string, string) {
	/* Try first common names */
	img := e.Data("cover.jpg")
	if len(img) != 0 {
		return storeImg(img, title, ".jpg")
	}
	img = e.Data("cover.jpeg")
	if len(img) != 0 {
		return storeImg(img, title, ".jpg")
	}
	img = e.Data("cover1.jpg")
	if len(img) != 0 {
		return storeImg(img, title, ".jpg")
	}
	img = e.Data("cover1.jpeg")
	if len(img) != 0 {
		return storeImg(img, title, ".jpg")
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

func parseFile(coll *mgo.Collection, path string) (string, error) {
	book := map[string]interface{}{}

	e, err := epub.Open(path, 0)
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

	coll.Insert(book)
	return title, nil
}

type uploadData struct {
	S Status
}

func uploadHandler(coll *mgo.Collection) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			sess := GetSession(r)
			paths, err := storeFiles(r)
			if err != nil {
				sess.Notify("Problem uploading!", "Some files were not stored. Try again or contact us if it keeps happening", "error")
			}

			uploaded := ""
			for _, path := range paths {
				title, err := parseFile(coll, path)
				if err != nil {
					os.Remove(path)
					sess.Notify("Problem uploading!", "The file '"+path[len("new/"):]+"' is not a well formed epub", "error")
				} else {
					uploaded = uploaded + " '" + title + "'"
				}
			}
			if uploaded != "" {
				sess.Notify("Upload successful!", "Added the books:"+uploaded+". Thank you for your contribution", "success")
			}
		}

		var data uploadData
		data.S = GetStatus(w, r)
		data.S.Upload = true
		loadTemplate(w, "upload", data)
	}
}
