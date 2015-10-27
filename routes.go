package main

import (
	"io"
	"log"
	"time"
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
	temps.ExecuteTemplate(w, "index", nil) // only serve index.html
}

func LoginTemplate(w http.ResponseWriter, r *http.Request) {
	temps.ExecuteTemplate(w, "login", nil)
}

func SignupTemplate(w http.ResponseWriter, r *http.Request) {
	temps.ExecuteTemplate(w, "signup", nil)
}

func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	switch {
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
			"Error":"Invalid Username",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
		}
		return
	case f["password"] == nil, len(f["password"]) != 1, f["password"][0] == "":
		err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
			"Error":"Invalid Password",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
		}
		return
	default:
		result := user{}
		err = muser.Find(bson.M{"username": f["username"][0]}).One(&result)
		switch err {
		case nil:
			break
		case mgo.ErrNotFound:
			err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
			}
			return
		default:
			log.Panic(err)
			return
		}
		// user found
		err = bcrypt.CompareHashAndPassword([]byte(result.Hash),
			[]byte(f["password"][0]))
		switch err {
		case nil:
			break
		case bcrypt.ErrMismatchedHashAndPassword:
			err = temps.ExecuteTemplate(w, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
			}
			return
		default:
			log.Panic(err)
			return
		}
		io.WriteString(w, "You are "+f["username"][0])
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	switch {
	// if username isn't present or not username field(s) or username is blank
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Bad Username",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
	// if password isn't present or not 2 password field(s) or password is blank
	case f["password"] == nil, len(f["password"]) != 2, f["password"][0] == "":
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Bad Password",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
	// if email isn't present or there aren't 1 email field(s) or email is blank
	case f["email"] == nil, len(f["email"]) != 1, f["email"][0] == "":
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Bad Email",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
	// if the two passwords don't match
	case f["password"][0] != f["password"][1]:
		err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Passwords do not match",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
	// otherwise regester user
	default:
		// result := user{}
		var i int
		i, err = muser.Find(bson.M{"username": f["username"][0]}).Count()
		if i < 1 {
			answer, err := bcrypt.GenerateFromPassword(
				[]byte(f["password"][0]), bcryptStrength)
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
				return
			}
			err = muser.Insert(&user{
				Id: bson.NewObjectId(),
				Username: f["username"][0],
				Hash: string(answer),
				Email: f["email"][0],
				Joined: time.Now(),
			})
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
				io.WriteString(w, "There was an error... Where did it go?")
				return
			}
			io.WriteString(w, "Thanks for signing up!")
			return // stop
		}else{
			err = temps.ExecuteTemplate(w, "signup", map[string]interface{}{
				"Error":"Username taken",
			})
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
				return
			}
		}
	}
}

func GetPeeves(c web.C, w http.ResponseWriter, r *http.Request) {
	user := user{}
	peeves := []peeve{}
	err = getUser(&user, c.URLParams["username"])
	switch err {
	case nil:
		break
	case mgo.ErrNotFound:
		err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
			"Number":"404",
			"Body":"Not Found",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
	default:
		log.Panic(err)
		return
	}
	err = getPeeves(&peeves, user.Id)
	switch err {
	case nil:
		break
	default:
		log.Panic(err)
		return
	}
	if len(peeves) > 0 {
		// if peeves
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Peeves": peeves,
			"User": user,
		})
	}else{
		// if no peeves
		err = temps.ExecuteTemplate(w, "user", nil)
	}
	switch err {
	case nil:
		break
	default:
		log.Panic(err)
		return // stop
	}
}

func CreatePeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	case f["body"]==nil, len(f["body"]) != 1, f["body"][0] == "":
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			// "Peeves": peeves,
			"Error": "Invalid Body",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
		}
		return // stop
	case len(f["body"][0]) > 140:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			// "Peeves": peeves,
			"Error": "Peeve Too Long",
		})
		switch err {
		case nil:
			break
		default:
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
				"Number":"404",
				"Body":"Not Found",
			})
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
			}
			return // stop
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(err)
			return
		}
		peeves := []peeve{}
		err = getPeeves(&peeves, user.Id)
		switch err {
		case nil:
			break
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(err)
			return // stop
		}
		mpeeve.Insert(&peeve{
			Id: bson.NewObjectId(),
			Creator: user.Id,
			User: user.Id, // create a peeve == owner
			Body: f["body"][0],
		})
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
	}
}

func DeletePeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	case f["id"] == nil, len(f["id"]) != 1,
		f["id"][0] == "", len(f["id"][0]) != 24:
		err = temps.ExecuteTemplate(w, "user", map[string]interface{}{
			"Error": "Invalid Id",
		})
		switch err {
		case nil:
			break
		default:
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
				"Number":"404",
				"Body":"Not Found",
			})
			switch err {
			case nil:
				break
			default:
				log.Panic(err)
				return // stop
			}
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(err)
			return
		}
		peeves := []peeve{}
		err = getPeeves(&peeves, user.Id)
		switch err {
		case nil:
			break
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(err)
			return // stop
		}
		err = dropPeeve(f["id"][0], user.Id)
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
	}
}

func Search(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// TODO: actually search something instead of just redirect to <user>
	f := r.Form
	switch {
	case f["q"]==nil, len(f["q"]) != 1, f["q"][0] == "":
		err = temps.ExecuteTemplate(w, "error", map[string]interface{}{
			"Number":"404",
			"Body":"Not Found",
		})
		switch err {
		case nil:
			break
		default:
			log.Panic(err)
			return
		}
	default:
		http.Redirect(w, r, "/"+f["q"][0], 302)
	}
}
