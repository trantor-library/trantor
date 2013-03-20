package main

import (
	"bytes"
	"git.gitorious.org/go-pkg/epub.git"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"log"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

func GetCover(e *epub.Epub, title string) (string, string) {
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

func storeImg(img []byte, title, extension string) (string, string) {
	r, _ := utf8.DecodeRuneInString(title)
	folder := string(r)
	if _, err := os.Stat(COVER_PATH + folder); err != nil {
		err = os.Mkdir(COVER_PATH+folder, os.ModePerm)
		if err != nil {
			log.Println("Error creating", COVER_PATH+folder, ":", err.Error())
			return "", ""
		}
	}

	/* open the files */
	imgPath := validFileName(COVER_PATH, title, extension)
	fBig, err := os.Create(COVER_PATH + imgPath)
	if err != nil {
		log.Println("Error creating", COVER_PATH+imgPath, ":", err.Error())
		return "", ""
	}
	defer fBig.Close()

	imgPathSmall := validFileName(COVER_PATH, title, "_small"+extension)
	fSmall, err := os.Create(COVER_PATH + imgPathSmall)
	if err != nil {
		log.Println("Error creating", COVER_PATH+imgPathSmall, ":", err.Error())
		return "", ""
	}
	defer fSmall.Close()

	/* resize img */
	jpgOptions := jpeg.Options{IMG_QUALITY}
	imgResized, err := resizeImg(img, IMG_WIDTH_BIG)
	if err != nil {
		log.Println("Error resizing big image:", err.Error())
		return "", ""
	}
	err = jpeg.Encode(fBig, imgResized, &jpgOptions)
	if err != nil {
		log.Println("Error encoding big image:", err.Error())
		return "", ""
	}
	imgSmallResized, err := resizeImg(img, IMG_WIDTH_SMALL)
	if err != nil {
		log.Println("Error resizing small image:", err.Error())
		return "", ""
	}
	err = jpeg.Encode(fSmall, imgSmallResized, &jpgOptions)
	if err != nil {
		log.Println("Error encoding small image:", err.Error())
		return "", ""
	}

	return imgPath, imgPathSmall
}

func resizeImg(imgBuff []byte, width uint) (image.Image, error) {
	reader := bytes.NewReader(imgBuff)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	return resize.Resize(width, 0, img, resize.NearestNeighbor), nil
}
