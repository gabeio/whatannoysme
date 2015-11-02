package main

import (
	"log"
	"time"
	"strings"
	"net/http"

	// golang.org/x/crypto
	"golang.org/x/crypto/bcrypt"

	// goji
	"github.com/zenazn/goji/web"

	// mgo
	"gopkg.in/mgo.v2/bson"
)

func CreateUser(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	switch {
	// if username isn't present or there aren't username field(s) or blank
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Bad Username",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	// if password isn't present or there aren't 2 password field(s) or blank
	case f["password"] == nil, len(f["password"]) != 2, f["password"][0] == "":
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Bad Password",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	// if email isn't present or there aren't 1 email field(s) or email is blank
	case f["email"] == nil, len(f["email"]) != 1, f["email"][0] == "":
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Bad Email",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	// max username length 13
	case len(f["username"][0]) > 13:
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Username too long",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	// if the two passwords don't match
	case f["password"][0] != f["password"][1]:
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Passwords do not match",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	}
	// otherwise regester user
	f["username"][0] = strings.ToLower(f["username"][0])
	i, err := muser.Find(bson.M{"username": f["username"][0]}).Count()
	if i > 1 {
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Username taken",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	}
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(f["password"][0]), bcryptStrength)
	if err != nil {
		log.Panic(err)
		return // stop
	}
	go createUser(&user{
		Id: bson.NewObjectId(),
		Username: f["username"][0],
		Hash: string(hash),
		Email: f["email"][0],
		Joined: time.Now(),
	}, errs)
	if <-errs != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Panic(<-errs)
		return // stop
	}
	session.Values["username"] = f["username"][0]
	session.Values["hash"] = string(hash)
	if err = session.Save(r, w); err != nil {
		log.Panic("Error saving session: %v", err)
	}
	http.Redirect(w, r, "/"+f["username"][0], 302)
	return // stop
}

func Settings(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != c.URLParams["username"] {
		// user is not this user
		http.Redirect(w, r, "/"+c.URLParams["username"]+"/settings", 302)
		return // stop
	}
	user := user{}
	go getUser(c.URLParams["username"], &user, errs)
	if <-errs != nil {
		log.Panic(<-errs)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	if len(f) > 0 {
		if len(f["first"]) == 1 {
			user.setFirstName(f["first"][0])
		}
		if len(f["last"]) == 1 {
			user.setLastName(f["last"][0])
		}
		if len(f["password"]) == 2 && f["password"][0] == f["password"][1] {
			user.setPassword(f["password"][0])
		}
	}else{
		http.Error(w, http.StatusText(500), 500)
		return
	}
	http.Redirect(w, r, "/"+c.URLParams["username"]+"/settings", 302)
	return // stop
}
