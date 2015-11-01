package main

import (
	"log"
	"time"
	"net/http"

	// goji
	"github.com/zenazn/goji/web"

	// mgo
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func GetPeeves(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"]
	user := user{}
	peeves := []peeve{}
	go getUser(c.URLParams["username"], &user, errs)
	switch <-errs {
	case nil:
		break
	case mgo.ErrNotFound:
		err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
			"Number": "404",
			"Body": "Not Found",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	default:
		log.Panic(<-errs)
		return // stop
	}
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
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
		// continue
	}
	username, _ := session.Values["username"].(string)
	if username != c.URLParams["username"] {
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
		go getUser(c.URLParams["username"], &user, errs)
		switch <-errs {
		case nil:
			break
		case mgo.ErrNotFound:
			err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
				"Session": session,
			})
			if err != nil {
				log.Panic(err)
			}
			return // stop
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return // stop
		}
		go createPeeve(&peeve{
			Id: bson.NewObjectId(),
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
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
		// continue
	}
	username, _ := session.Values["username"].(string)
	if username != c.URLParams["username"] {
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 24:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid Id",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
			return // stop
		}
	default:
		user := user{}
		go getUser(c.URLParams["username"], &user, errs)
		switch <-errs {
		case nil:
			break
		case mgo.ErrNotFound:
			err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
				"Session": session,
			})
			if err != nil {
				log.Panic(err)
				return // stop
			}
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return // stop
		}
		go dropPeeve(bson.ObjectIdHex(f["id"][0]), errs)
		if <-errs != nil {
			log.Panic(<-errs)
			return // stop
		}
		http.Redirect(w, r, "/"+c.URLParams["username"], 302)
	}
}
