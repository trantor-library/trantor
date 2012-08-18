package main

import (
    "net/http"
    "code.google.com/p/gorilla/sessions"
    "code.google.com/p/gorilla/securecookie"
)

var sesStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))

func CreateSession(user string, w http.ResponseWriter, r *http.Request) {
	session, _ := sesStore.Get(r, "admin")
	session.Values["user"] = user
	session.Save(r, w)
}

func SessionUser(r *http.Request) string {
	session, err := sesStore.New(r, "admin")
	if err != nil {
		return ""
	}
	if session.IsNew {
		return ""
	}
	user, ok := session.Values["user"].(string)
	if !ok {
		return ""
	}
	return user
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	session, err := sesStore.Get(r, "admin")
	if err != nil {
		return
	}
	if session.IsNew {
		return
	}
	session.Values["user"] = ""
	session.Options.MaxAge = -1
	session.Save(r, w)
}
