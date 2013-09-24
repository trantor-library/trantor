package main

import (
	"encoding/hex"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
)

var sesStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(64))

type Notification struct {
	Title string
	Msg   string
	Type  string /* error, info or success */
}

type Session struct {
	User string
	Role string
	S    *sessions.Session
}

func GetSession(r *http.Request, db *DB) (s *Session) {
	s = new(Session)
	var err error
	s.S, err = sesStore.Get(r, "session")
	if err == nil && !s.S.IsNew {
		s.User, _ = s.S.Values["user"].(string)
		s.Role = db.UserRole(s.User)
	}

	if s.S.IsNew {
		s.S.Values["id"] = hex.EncodeToString(securecookie.GenerateRandomKey(16))
	}

	return
}

func (s *Session) GetNotif() []Notification {
	session := s.S
	msgs := session.Flashes("nMsg")
	titles := session.Flashes("nTitle")
	tpes := session.Flashes("nType")
	notif := make([]Notification, len(msgs))
	for i, m := range msgs {
		msg, _ := m.(string)
		title, _ := titles[i].(string)
		tpe, _ := tpes[i].(string)
		notif[i] = Notification{title, msg, tpe}
	}
	return notif
}

func (s *Session) LogIn(user string) {
	s.User = user
	s.S.Values["user"] = user
}

func (s *Session) LogOut() {
	s.S.Values["user"] = ""
}

func (s *Session) Notify(title, msg, tpe string) {
	s.S.AddFlash(msg, "nMsg")
	s.S.AddFlash(title, "nTitle")
	s.S.AddFlash(tpe, "nType")
}

func (s *Session) Save(w http.ResponseWriter, r *http.Request) {
	sesStore.Save(r, w, s.S)
}

func (s *Session) Id() string {
	id, _ := s.S.Values["id"].(string)
	return id
}

func (s *Session) IsAdmin() bool {
	return s.Role == "admin"
}
