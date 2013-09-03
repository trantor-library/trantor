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

func createUserHandler(w http.ResponseWriter, r *http.Request, sess *Session) {
	pass := r.FormValue("pass")
	confirmPass := r.FormValue("confirmPass")
	if pass != confirmPass {
		sess.Notify("Registration error!", "Passwords don't match", "error")
	} else {
		user := r.FormValue("user")
		err := db.AddUser(user, pass)
		if err == nil {
			sess.Notify("Account created!", "Welcome "+user+". Now you can login", "success")
		} else {
			sess.Notify("Registration error!", "There was some database problem, if it keeps happening please inform me", "error")
		}
	}
	sess.Save(w, r)
	http.Redirect(w, r, r.Referer(), http.StatusFound)
}
