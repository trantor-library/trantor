// +build prod

package main

import (
	"io/ioutil"
	"strings"

	"git.gitorious.org/go-pkg/epubgo.git"
	"github.com/rainycape/cld2"
)

func GuessLang(epub *epubgo.Epub, orig_langs []string) []string {
	spine, err := epub.Spine()
	if err != nil {
		return orig_langs
	}

	var err_spine error
	err_spine = nil
	langs := []string{}
	for err_spine == nil {
		html, err := spine.Open()
		err_spine = spine.Next()
		if err != nil {
			continue
		}
		defer html.Close()

		buff, err := ioutil.ReadAll(html)
		if err != nil {
			continue
		}
		langs = append(langs, cld2.Detect(string(buff)))
	}

	lang := commonLang(langs)
	if lang != "un" && differentLang(lang, orig_langs) {
		return []string{lang}
	}
	return orig_langs
}

func commonLang(langs []string) string {
	count := map[string]int{}
	for _, l := range langs {
		count[l]++
	}

	lang := "un"
	maxcount := 0
	for l, c := range count {
		if c > maxcount && l != "un" {
			lang = l
			maxcount = c
		}
	}
	return lang
}

func differentLang(lang string, orig_langs []string) bool {
	orig_lang := "un"
	if len(orig_langs) > 0 && len(orig_langs) >= 2 {
		orig_lang = strings.ToLower(orig_langs[0][0:2])
	}

	return orig_lang != lang
}
