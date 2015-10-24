package main

import (
	"io"
	"log"
	"time"
	"strconv"
	"net/http"

	// golang.org/crypto
	"golang.org/x/crypto/bcrypt"

	// goji
	"github.com/zenazn/goji/web"

	// mgo
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func Root(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "index", nil) // only serve index.html
}

func Login(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "login", nil)
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1MB limit
	f := r.Form
	switch {
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = tem.ExecuteTemplate(w, "login", map[string]string{"Error":"Invalid Username"})
	case f["password"] == nil, len(f["password"]) != 1, f["password"][0] == "":
		err = tem.ExecuteTemplate(w, "login", map[string]string{"Error":"Invalid Password"})
	default:
		result := User{}
		err = muser.Find(bson.M{"username": f["username"][0]}).One(&result)
		log.Print(result)
		if err != nil {
			if err == mgo.ErrNotFound {
				err = tem.ExecuteTemplate(w, "login", map[string]string{"Error":"Invalid Username or Password"})
				if err != nil {
					log.Panic(err)
				}
			}else{
				log.Panic(err)
			}
		}else{
			err = bcrypt.CompareHashAndPassword([]byte(result.Hash),[]byte(f["password"][0]))
			if err != nil {
				if err == bcrypt.ErrMismatchedHashAndPassword {
					err = tem.ExecuteTemplate(w, "login", map[string]string{"Error":"Invalid Username or Password"})
					if err != nil {
						log.Panic(err)
					}
				}else{
					log.Panic(err)
				}
			}else{
				io.WriteString(w, "You are "+f["username"][0])
			}
		}
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "signup", nil)
}

func PostSignup(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1MB limit
	f := r.Form
	switch {
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = tem.ExecuteTemplate(w, "signup", map[string]string{"Error":"Bad Username"})
		if err != nil {
			log.Panic(err)
		}
	case f["password"] == nil, len(f["password"]) != 2, f["password"][0] == "":
		err = tem.ExecuteTemplate(w, "signup", map[string]string{"Error":"Bad Password"})
		if err != nil {
			log.Panic(err)
		}
	case f["email"] == nil, len(f["email"]) != 1, f["email"][0] == "":
		err = tem.ExecuteTemplate(w, "signup", map[string]string{"Error":"Bad Email"})
		if err != nil {
			log.Panic(err)
		}
	case f["password"][0] != f["password"][1]:
		err = tem.ExecuteTemplate(w, "signup", map[string]string{"Error":"Passwords do not match"})
		if err != nil {
			log.Panic(err)
		}
	default:
		answer, err := bcrypt.GenerateFromPassword([]byte(f["password"][0]), 11)
		if err != nil {
			log.Panic(err)
		}else{
			log.Print(answer)
		}
		err = muser.Insert(&User{
			Id: bson.NewObjectId(),
			Username: f["username"][0],
			Hash: string(answer),
			Email: f["email"][0],
			Joined: time.Now(),
		});
		if err != nil {
			log.Panic(err)
			io.WriteString(w, "There was an error... Where did it get to?")
		}else{
			io.WriteString(w, "Thanks for signing up!")
		}
	}
}

func Random(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, strconv.FormatInt(ran.Int63(), 10))
}

func GetUser(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Print(c.URLParams["username"])
	tem.ExecuteTemplate(w, "user", c.URLParams["username"])
}

func PostUser(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1048576) // translate multipart
	if r.Form["body"] == nil {
		// log.Print(w.Body)
		// log.Print(r)
		// w.Body.Close()
		io.WriteString(w,"") // closest thing to .close()
	}
	io.WriteString(w,r.Form["body"][0])
}
