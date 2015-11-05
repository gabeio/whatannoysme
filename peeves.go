package main

import (
	"log"
	"time"
	"net/http"

	// goji
	"github.com/zenazn/goji/web"

	// rethink
	"github.com/dancannon/gorethink"
)

func GetPeeves(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	user := user{}
	go getOneUser(c.URLParams["username"], &user, errs)
	switch <-errs {
	case nil:
		break
	case gorethink.ErrEmptyResult:
		err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
			"Number": "404",
			"Body": "Not Found",
			"SessionUsername": username, // this might be blank
			"Session": session, // this might be blank
		})
		if err != nil {
			log.Panic(err)
		}
		return
	default:
		log.Panic(<-errs)
		return // stop
	}
	peeves := []peeve{}
	go getPeeves(user.Id, &peeves, errs)
	if <-errs != nil {
		log.Panic(<-errs)
		return // stop
	}
	err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
		"Peeves": peeves,
		"User": user,
		"SessionUsername": username,
		"Session": session,
	})
	if err != nil {
		log.Panic(err)
		return // stop
	}
}

func CreatePeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.URLParams["username"] { // if user logged isn't this user
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	case f["body"] == nil, len(f["body"]) != 1, f["body"][0] == "":
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid Body",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	case len(f["body"][0]) > 140:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Peeve Too Long",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	default:
		user := user{}
		go getOneUser(c.URLParams["username"], &user, errs)
		if <-errs != nil {
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return // stop
		}
		go createPeeve(&peeve{
			// Id: bson.NewObjectId(),
			Root: user.Id,
			// as this is the root no parent
			UserId: user.Id, // create a peeve == owner
			Body: f["body"][0],
			Timestamp: time.Now(),
		}, errs)
		if <-errs != nil {
			log.Panic(<-errs)
		}
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
	}
}

func DeletePeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.URLParams["username"] { // if user logged isn't this user
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid Id",
			// "Peeves": peeves,
			// "User": user,
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
			return // stop
		}
	default:
		user := user{}
		go getOneUser(username, &user, errs)
		if <-errs != nil{
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return // stop
		}
		go dropOnePeeve(f["id"][0], user.Id, errs)
		if <-errs != nil {
			log.Panic(<-errs)
			return // stop
		}
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
	}
}

func MeTooPeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username == c.URLParams["username"] {
		// don't allow a user to metoo their own peeve
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid Id",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
			return // stop
		}
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["user"] == nil, len(f["user"]) != 1, len(f["user"][0]) != 36:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid User",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	default:
		user := user{}
		go getOneUser(username, &user, errs)
		switch <-errs {
		case nil:
			break
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return // stop
		}
		metoopeeve := peeve{}
		// don't assume input is valid
		go getOnePeeve(f["id"][0], f["user"][0], &metoopeeve, errs)
		switch <-errs {
		case nil:
			break
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return // stop
		}
		peevey := &peeve{
			// Id: bson.NewObjectId(), // create new id
			Root: metoopeeve.Root, // peeve origin
			Parent: metoopeeve.UserId, // who I got it from
			UserId: user.Id, // who owns this peeve
			Body: metoopeeve.Body, // original body
			Timestamp: time.Now(), // when I reposted it
		}
		go createPeeve(peevey, errs)
		if <-errs != nil {
			log.Panic(<-errs)
		}
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
	}
}
