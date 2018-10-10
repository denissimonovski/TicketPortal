package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type userticket struct {
	Emuser User
	Tiketi []tiket
}

type tiket struct {
	Id                                                      int
	Pusten_od, Go_raboti, Otvoren, First_response, Zatvoren string
	//Otvoren, First_response, Zatvoren time.Time
}

type User struct {
	Un, Fn, Ln string
	Ps         []byte
}

var tpl *template.Template
var sesie = map[string]string{}
var korisnici = map[string]User{}
var db *sql.DB
var err error

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

func main() {
	db, err = sql.Open("mysql", "root:Kumanovo123$@tcp("+
		"mysql:3306)/cases?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", index)
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/read", read)
	http.HandleFunc("/login", login)
	http.HandleFunc("/inside", inside)
	http.HandleFunc("/logout", logout)
	http.Handle("/favicon.ico", http.NotFoundHandler())

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func index(w http.ResponseWriter, req *http.Request) {
	u := getUser(w, req)
	tpl.ExecuteTemplate(w, "index.gohtml", u)
}

func signup(w http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {

		un := req.FormValue("un")
		if _, ok := korisnici[un]; ok {
			http.Error(w, "Username already exists", http.StatusForbidden)
			return
		}
		ps, e := bcrypt.GenerateFromPassword([]byte(req.FormValue("ps")),
			bcrypt.MinCost)
		if e != nil {
			log.Fatal(e)
		}
		u := User{
			Un: un,
			Ps: ps,
			Fn: req.FormValue("fn"),
			Ln: req.FormValue("ln"),
		}

		c, e := req.Cookie("sessions")

		if e == http.ErrNoCookie {
			uid, _ := uuid.NewV4()
			c = &http.Cookie{
				Name:  "sessions",
				Value: uid.String(),
			}
			http.SetCookie(w, c)
		}

		sesie[c.Value] = un
		korisnici[un] = u

		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "signup.gohtml", nil)
}

func login(w http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	var incorrect bool = false
	if req.Method == http.MethodPost {
		un := req.FormValue("un")
		ps := req.FormValue("ps")
		if u, ok := korisnici[un]; ok {
			if bcrypt.CompareHashAndPassword(u.Ps, []byte(ps)) == nil {
				c, e := req.Cookie("sessions")
				if e == http.ErrNoCookie {
					uid, _ := uuid.NewV4()

					c = &http.Cookie{
						Name:  "sessions",
						Value: uid.String(),
					}
					http.SetCookie(w, c)
				}
				sesie[c.Value] = un

				http.Redirect(w, req, "/", http.StatusSeeOther)
			} else {
				incorrect = true
			}
		}
	}

	tpl.ExecuteTemplate(w, "login.gohtml", incorrect)
}

func inside(w http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	u := getUser(w, req)

	stmt, e := db.Prepare(`SELECT * FROM zapisi;`)
	checkErr(e)
	defer stmt.Close()

	tiketi := []tiket{}
	rows, er := stmt.Query()
	checkErr(er)
	var id int
	var pusten, raboti, otvoren, first_response, zatvoren string

	inlayout := "2006-01-02 15:04:05"
	outlayout := "Mon, Jan 2, 15:04:05 MST 2006"
	loc, _ := time.LoadLocation("Europe/Skopje")
	for rows.Next() {
		err = rows.Scan(&id,
			&pusten,
			&raboti,
			&otvoren,
			&first_response,
			&zatvoren)
		checkErr(err)
		otv, _ := time.ParseInLocation(inlayout, otvoren, loc)
		fr, _ := time.ParseInLocation(inlayout, first_response, loc)
		ztv, _ := time.ParseInLocation(inlayout, zatvoren, loc)
		tkt := tiket{
			Id:             id,
			Pusten_od:      pusten,
			Go_raboti:      raboti,
			Otvoren:        otv.Format(outlayout),
			First_response: fr.Format(outlayout),
			Zatvoren:       ztv.Format(outlayout),
		}
		tiketi = append(tiketi, tkt)
	}
	ut := userticket{Emuser: u, Tiketi: tiketi}
	tpl.ExecuteTemplate(w, "inside.gohtml", ut)
}

func logout(w http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}
	c, _ := req.Cookie("sessions")
	delete(sesie, c.Value)
	c = &http.Cookie{
		Name:   "sessions",
		Value:  "",
		MaxAge: -1,
	}

	http.SetCookie(w, c)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}
