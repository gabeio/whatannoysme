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
)

func CreateUser(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
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
	if len(strings.Fields(f["username"][0])) > 1 {
		// username has \t \n or space
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error": "Username Contains Invalid Characters",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	}
	// force all usernames to be lowercase
	f["username"][0] = strings.ToLower(f["username"][0])
	var i int
	go getCountUsername(f["username"][0], &i, errs)
	if <-errs != nil {
		log.Panic(<-errs)
	}
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
	newuser := &user{
		// Id: bson.NewObjectId(),
		Username: f["username"][0],
		Hash: string(hash),
		Email: f["email"][0],
		Joined: time.Now(),
	}
	go createUser(newuser, errs)
	if <-errs != nil {
		http.Error(w, http.StatusText(500), 500)
		log.Panic(<-errs)
		return // stop
	}
	// session.Values["user"] = user
	session.Values["username"] = newuser.Username
	session.Values["hash"] = string(hash)
	if err = session.Save(r, w); err != nil {
		log.Panic("Error saving session: %v", err)
	}
	http.Redirect(w, r, "/"+f["username"][0], 302)
	return // stop
}

func Login(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
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
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
			"Error": "Invalid Username",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	case f["password"] == nil, len(f["password"]) != 1, f["password"][0] == "":
		err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
			"Error": "Invalid Password",
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	default:
		f["username"][0] = strings.ToLower(f["username"][0]) // assure one user per username
		user := user{}
		go getOneUser(f["username"][0], &user, errs)
		switch <-errs {
		case nil:
			break
		default:
			log.Panic(<-errs)
			return // stop
		}
		// user found
		err = bcrypt.CompareHashAndPassword([]byte(user.Hash),
			[]byte(f["password"][0]))
		switch err {
		case nil:
			break
		case bcrypt.ErrMismatchedHashAndPassword:
			// incorrect password
			err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			if err != nil {
				log.Panic(err)
			}
			return // stop
		default:
			log.Panic(err)
			return // stop
		}
		// correct password
		// session.Values["user"] = user
		session.Values["username"] = user.Username
		session.Values["hash"] = user.Hash
		if err = session.Save(r, w); err != nil {
			log.Panic("Error saving session: %v", err)
		}
		http.Redirect(w, r, "/"+f["username"][0], 302)
	}
}

func Logout(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		log.Panic("Error saving session: %v", err)
	}
	http.Redirect(w, r, "/", 302)
}

func Settings(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != c.URLParams["username"] {
		// user is not this user
		http.Redirect(w, r, "/"+c.URLParams["username"]+"/settings", 302)
		return // stop
	}
	thisuser := user{}
	go getOneUser(c.URLParams["username"], &thisuser, errs)
	if <-errs != nil {
		log.Panic(<-errs)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	if len(f) > 0 {
		if len(f["first"]) == 1 {
			thisuser.setFirstName(f["first"][0])
		}
		if len(f["last"]) == 1 {
			thisuser.setLastName(f["last"][0])
		}
		if len(f["password"]) == 2 && f["password"][0] == f["password"][1] {
			thisuser.setPassword(f["password"][0])
		}
	}else{
		http.Error(w, http.StatusText(500), 500)
		return
	}
	http.Redirect(w, r, "/"+c.URLParams["username"]+"/settings", 302)
	return // stop
}
