package main

import (
	"fmt"
	"github.com/satori/go.uuid"
	"net/http"
)

func read(w http.ResponseWriter, req *http.Request) {
	uid := req.URL.RawQuery
	if un, ok := sesie[uid]; ok {
		fmt.Fprint(w, un)
	} else {
		fmt.Fprint(w, "Username doesn't exist")
	}
}

func getUser(w http.ResponseWriter, r *http.Request) User {
	var u User

	c, e := r.Cookie("sessions")
	if e == http.ErrNoCookie {

		uid, _ := uuid.NewV4()

		c = &http.Cookie{
			Name:  "sessions",
			Value: uid.String(),
		}
		http.SetCookie(w, c)
		return u
	}

	if un, ok := sesie[c.Value]; ok {
		return korisnici[un]
	}

	return u
}

func alreadyLoggedIn(r *http.Request) bool {
	c, e := r.Cookie("sessions")
	if e == http.ErrNoCookie {
		return false
	}
	un, ok := sesie[c.Value]
	if ok {
		_, uok := korisnici[un]
		return uok
	}
	return false
}

func checkErr(e error) {
	if e != nil {
		fmt.Println(e)
	}
}