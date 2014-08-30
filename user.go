package main

import (
	log "github.com/cihub/seelog"

	"net/http"
)

func loginHandler(h handler) {
	if h.sess.User != "" {
		http.Redirect(h.w, h.r, "/dashboard/", http.StatusFound)
		return
	}

	var data statusData
	data.S = GetStatus(h)
	loadTemplate(h, "login", data)
}

func loginPostHandler(h handler) {
	user := h.r.FormValue("user")
	pass := h.r.FormValue("pass")
	if h.db.User(user).Valid(pass) {
		log.Info("User ", user, " log in")
		h.sess.LogIn(user)
		h.sess.Notify("Successful login!", "Welcome "+user, "success")
	} else {
		log.Warn("User ", user, " bad user or password")
		h.sess.Notify("Invalid login!", "user or password invalid", "error")
	}
	h.sess.Save(h.w, h.r)
	http.Redirect(h.w, h.r, h.r.Referer(), http.StatusFound)
}

func createUserHandler(h handler) {
	pass := h.r.FormValue("pass")
	confirmPass := h.r.FormValue("confirmPass")
	if pass != confirmPass {
		h.sess.Notify("Registration error!", "Passwords don't match", "error")
	} else {
		user := h.r.FormValue("user")
		err := h.db.AddUser(user, pass)
		if err == nil {
			h.sess.Notify("Account created!", "Welcome "+user+". Now you can login", "success")
		} else {
			h.sess.Notify("Registration error!", "There was some database problem, if it keeps happening please inform me", "error")
		}
	}
	h.sess.Save(h.w, h.r)
	http.Redirect(h.w, h.r, h.r.Referer(), http.StatusFound)
}

func dashboardHandler(h handler) {
	if h.sess.User == "" {
		notFound(h)
		return
	}

	var data statusData
	data.S = GetStatus(h)
	data.S.Dasboard = true
	loadTemplate(h, "dashboard", data)
}

func settingsHandler(h handler) {
	if h.sess.User == "" {
		notFound(h)
		return
	}
	if h.r.Method == "POST" {
		current_pass := h.r.FormValue("currpass")
		pass1 := h.r.FormValue("password1")
		pass2 := h.r.FormValue("password2")
		switch {
		case !h.db.User(h.sess.User).Valid(current_pass):
			h.sess.Notify("Password error!", "The current password given don't match with the user password. Try again", "error")
		case pass1 != pass2:
			h.sess.Notify("Passwords don't match!", "The new password and the confirmation password don't match. Try again", "error")
		default:
			h.db.User(h.sess.User).SetPassword(pass1)
			h.sess.Notify("Password updated!", "Your new password is correctly set.", "success")
		}
		h.sess.Save(h.w, h.r)
	}

	var data statusData
	data.S = GetStatus(h)
	loadTemplate(h, "settings", data)
}
