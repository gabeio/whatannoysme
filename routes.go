package main

import (
	"io"
	"log"
	"time"
	"strings"
	"net/http"

	// golang.org/crypto
	"golang.org/x/crypto/bcrypt"

	// goji
	"github.com/zenazn/goji/web"

	// mgo
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func IndexTemplate(w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302) // redirect to their page
		return // stop
	}
	temps.ExecuteTemplate(w, "index", nil) // only serve index.html
}

func SignupTemplate(w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302)
		return // stop
	}
	temps.ExecuteTemplate(w, "signup", nil)
}

func LoginTemplate(w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302)
		return // stop
	}
	temps.ExecuteTemplate(w, "login", nil)
}

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
	// otherwise regester user
	default:
		f["username"][0] = strings.ToLower(f["username"][0])
		var i int
		i, err = muser.Find(bson.M{"username": f["username"][0]}).Count()
		if i < 1 {
			hash, err := bcrypt.GenerateFromPassword(
				[]byte(f["password"][0]), bcryptStrength)
			if err != nil {
				log.Panic(err)
				return // stop
			}
			err = muser.Insert(&user{
				Id: bson.NewObjectId(),
				Username: f["username"][0],
				Hash: string(hash),
				Email: f["email"][0],
				Joined: time.Now(),
			})
			if err != nil {
				log.Panic(err)
				io.WriteString(w, "There was an error... Where did it go?")
				return // stop
			}
			session.Values["username"] = f["username"][0]
			session.Values["hash"] = string(hash)
			if err = session.Save(r, w) ; err != nil {
				log.Panic("Error saving session: %v", err)
			}
			http.Redirect(w, r, "/"+f["username"][0], 302)
			return // stop
		}else{
			err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
				"Error": "Username taken",
			})
			if err != nil {
				log.Panic(err)
			}
			return // stop
		}
	}
}

func Login(c web.C, w http.ResponseWriter, r *http.Request) {
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
		f["username"][0] = strings.ToLower(f["username"][0])
		user := user{}
		err = muser.Find(bson.M{"username": f["username"][0]}).One(&user)
		switch err {
		case nil:
			break
		case mgo.ErrNotFound:
			// user not found
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
		session.Values["username"] = f["username"][0]
		session.Values["hash"] = user.Hash
		if err = session.Save(r, w) ; err != nil {
			log.Panic("Error saving session: %v", err)
		}
		http.Redirect(w,r,"/"+f["username"][0],302)
	}
}

func Logout(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(r, w); err != nil {
		log.Panic("Error saving session: %v", err)
	}
	http.Redirect(w,r,"/",302)
}

func GetPeeves(c web.C, w http.ResponseWriter, r *http.Request) {
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"]
	user := user{}
	peeves := []peeve{}
	err = getUser(&user, c.URLParams["username"])
	switch err {
	case nil:
		break
	case mgo.ErrNotFound:
		err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
			"Number": "404",
			"Body": "Not Found",
			"SessionUsername": username,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	default:
		log.Panic(err)
		return // stop
	}
	err = getPeeves(&peeves, user.Id)
	if err != nil {
		log.Panic(err)
		return // stop
	}
	err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
		"Peeves": peeves,
		"User": user,
		"SessionUsername": username,
	})
	if err != nil {
		log.Panic(err)
		return // stop
	}
}

func CreatePeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	var username string
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	switch v := session.Values["username"].(type) {
	case string:
		if v != c.URLParams["username"]{
			http.Redirect(w,r,"/"+c.URLParams["username"],302)
			return // stop
		}else{
			username = v
			break
		}
	default:
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
		return // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	case f["body"]==nil, len(f["body"]) != 1, f["body"][0] == "":
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid Body",
			"SessionUsername": username,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	case len(f["body"][0]) > 140:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Peeve Too Long",
			"SessionUsername": username,
		})
		if err != nil {
			log.Panic(err)
		}
		return // stop
	default:
		user := user{}
		err = getUser(&user, c.URLParams["username"])
		switch err {
		case nil:
			break
		case mgo.ErrNotFound:
			err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
			})
			if err != nil {
				log.Panic(err)
			}
			return // stop
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(err)
			return // stop
		}
		peeves := []peeve{}
		err = getPeeves(&peeves, user.Id)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			log.Panic(err)
			return // stop
		}
		err = mpeeve.Insert(&peeve{
			Id: bson.NewObjectId(),
			Root: user.Id,
			// as this is the root no parent
			User: user.Id, // create a peeve == owner
			Body: f["body"][0],
			Timestamp: time.Now(),
		})
		if err != nil {
			log.Panic(err)
		}
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
	}
}

func DeletePeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	var username string
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	switch v := session.Values["username"].(type) {
	case string:
		if v != c.URLParams["username"]{
			// user is trying to delete a peeve which they do not own
			http.Redirect(w,r,"/"+c.URLParams["username"],302)
			return // stop
		}else{
			username = v
			break
		}
	default:
		// unknown is trying to delete a peeve which they do not own
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
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
		err = getUser(&user, c.URLParams["username"])
		switch err {
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
			log.Panic(err)
			return // stop
		}
		err = dropPeeve(f["id"][0], user.Id)
		if err != nil {
			log.Panic(err)
			return // stop
		}
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
	}
}

func Search(w http.ResponseWriter, r *http.Request) {
	var username string
	session, err := rstore.Get(r, "wam")
	if err != nil {
		log.Panic(err)
	}
	username, _ = session.Values["username"].(string)
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// TODO: actually search something instead of just redirect to <user>
	f := r.Form
	switch {
	case f["q"]==nil, len(f["q"]) != 1, f["q"][0] == "":
		err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
			"Number": "404",
			"Body": "Not Found",
			"SessionUsername": username,
		})
		if err != nil {
			log.Panic(err)
			return // stop
		}
	default:
		http.Redirect(w, r, "/"+f["q"][0], 302)
	}
}
