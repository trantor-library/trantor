// +build !prod

// This is a dummy implementation of GuessLang used to make the compilation faster on development
//
// To build trantor with the proper language guessing do:
//   $ go build -tags prod

package main

import (
	"git.gitorious.org/go-pkg/epubgo.git"
)

func GuessLang(epub *epubgo.Epub, orig_langs []string) []string {
	return orig_langs
}
