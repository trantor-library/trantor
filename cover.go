package main

import log "github.com/cihub/seelog"
import _ "image/png"
import _ "image/jpeg"
import _ "image/gif"

import (
	"bytes"
	"git.gitorious.org/go-pkg/epubgo.git"
	"git.gitorious.org/trantor/trantor.git/storage"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

func coverHandler(h handler) {
	vars := mux.Vars(h.r)
	book, err := h.db.GetBookId(vars["id"])
	if err != nil {
		notFound(h)
		return
	}

	if !book.Active {
		if !h.sess.IsAdmin() {
			notFound(h)
			return
		}
	}

	file := COVER_FILE
	if vars["size"] == "small" {
		file = COVER_SMALL_FILE
	}
	f, err := h.store.Get(book.Id, file)
	if err != nil {
		log.Error("Error while opening image: ", err)
		notFound(h)
		return
	}
	defer f.Close()

	headers := h.w.Header()
	headers["Content-Type"] = []string{"image/jpeg"}

	_, err = io.Copy(h.w, f)
	if err != nil {
		log.Error("Error while copying image: ", err)
		notFound(h)
		return
	}
}

func GetCover(e *epubgo.Epub, id string, store *storage.Store) bool {
	if coverFromMetadata(e, id, store) {
		return true
	}

	if searchCommonCoverNames(e, id, store) {
		return true
	}

	/* search for img on the text */
	exp, _ := regexp.Compile("<.*ima?g.*[(src)(href)]=[\"']([^\"']*(\\.[^\\.\"']*))[\"']")
	it, errNext := e.Spine()
	for errNext == nil {
		file, err := it.Open()
		if err != nil {
			break
		}
		defer file.Close()

		txt, err := ioutil.ReadAll(file)
		if err != nil {
			break
		}
		res := exp.FindSubmatch(txt)
		if res != nil {
			href := string(res[1])
			urlPart := strings.Split(it.URL(), "/")
			url := strings.Join(urlPart[:len(urlPart)-1], "/")
			if href[:3] == "../" {
				href = href[3:]
				url = strings.Join(urlPart[:len(urlPart)-2], "/")
			}
			href = strings.Replace(href, "%20", " ", -1)
			href = strings.Replace(href, "%27", "'", -1)
			href = strings.Replace(href, "%28", "(", -1)
			href = strings.Replace(href, "%29", ")", -1)
			if url == "" {
				url = href
			} else {
				url = url + "/" + href
			}

			img, err := e.OpenFile(url)
			if err == nil {
				defer img.Close()
				return storeImg(img, id, store)
			}
		}
		errNext = it.Next()
	}
	return false
}

func coverFromMetadata(e *epubgo.Epub, id string, store *storage.Store) bool {
	metaList, _ := e.MetadataAttr("meta")
	for _, meta := range metaList {
		if meta["name"] == "cover" {
			img, err := e.OpenFileId(meta["content"])
			if err == nil {
				defer img.Close()
				return storeImg(img, id, store)
			}
		}
	}
	return false
}

func searchCommonCoverNames(e *epubgo.Epub, id string, store *storage.Store) bool {
	for _, p := range []string{"cover.jpg", "Images/cover.jpg", "images/cover.jpg", "cover.jpeg", "cover1.jpg", "cover1.jpeg"} {
		img, err := e.OpenFile(p)
		if err == nil {
			defer img.Close()
			return storeImg(img, id, store)
		}
	}
	return false
}

func storeImg(img io.Reader, id string, store *storage.Store) bool {
	/* open the files */
	fBig, err := store.Create(id, COVER_FILE)
	if err != nil {
		log.Error("Error creating cover ", id, ": ", err.Error())
		return false
	}
	defer fBig.Close()

	fSmall, err := store.Create(id, COVER_SMALL_FILE)
	if err != nil {
		log.Error("Error creating small cover ", id, ": ", err.Error())
		return false
	}
	defer fSmall.Close()

	/* resize img */
	var img2 bytes.Buffer
	img1 := io.TeeReader(img, &img2)
	jpgOptions := jpeg.Options{IMG_QUALITY}
	imgResized, err := resizeImg(img1, IMG_WIDTH_BIG)
	if err != nil {
		log.Error("Error resizing big image: ", err.Error())
		return false
	}
	err = jpeg.Encode(fBig, imgResized, &jpgOptions)
	if err != nil {
		log.Error("Error encoding big image: ", err.Error())
		return false
	}
	imgSmallResized, err := resizeImg(&img2, IMG_WIDTH_SMALL)
	if err != nil {
		log.Error("Error resizing small image: ", err.Error())
		return false
	}
	err = jpeg.Encode(fSmall, imgSmallResized, &jpgOptions)
	if err != nil {
		log.Error("Error encoding small image: ", err.Error())
		return false
	}
	return true
}

func resizeImg(imgReader io.Reader, width uint) (image.Image, error) {
	img, _, err := image.Decode(imgReader)
	if err != nil {
		return nil, err
	}

	return resize.Resize(width, 0, img, resize.NearestNeighbor), nil
}
