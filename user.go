package main

import (
	"log"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	var data statusData
	data.S = GetStatus(w, r)
	loadTemplate(w, "login", data)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	if db.UserValid(user, pass) {
		log.Println("User", user, "log in")
		sess.LogIn(user)
		sess.Notify("Successful login!", "Welcome "+user, "success")
	} else {
		log.Println("User", user, "bad user or password")
		sess.Notify("Invalid login!", "user or password invalid", "error")
	}
	sess.Save(w, r)
	http.Redirect(w, r, r.Referer(), http.StatusFound)
}
