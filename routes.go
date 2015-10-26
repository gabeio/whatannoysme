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
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	switch {
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = tem.ExecuteTemplate(w, "login", map[string]interface{}{
			"Error":"Invalid Username",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	case f["password"] == nil, len(f["password"]) != 1, f["password"][0] == "":
		err = tem.ExecuteTemplate(w, "login", map[string]interface{}{
			"Error":"Invalid Password",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	default:
		result := User{}
		err = muser.Find(bson.M{"username": f["username"][0]}).One(&result)
		if err != nil {
			// mgo error
			if err == mgo.ErrNotFound { // user not found
				err = tem.ExecuteTemplate(w, "login", map[string]interface{}{
					"Error": "Invalid Username or Password",
				})
				if err != nil {
					// html/template error
					log.Panic(err)
				}
				return // stop
			}
			// otherwise
			log.Panic(err)
			return // stop
		}
		// user found
		err = bcrypt.CompareHashAndPassword([]byte(result.Hash),
			[]byte(f["password"][0]))
		if err != nil {
			// bcrypt error
			if err == bcrypt.ErrMismatchedHashAndPassword { // wrong password
				err = tem.ExecuteTemplate(w, "login", map[string]interface{}{
					"Error": "Invalid Username or Password",
				})
				if err != nil {
					// html/template error
					log.Panic(err)
				}
				return // stop
			}
			// otherwise
			log.Panic(err)
			return // stop
		}
		io.WriteString(w, "You are "+f["username"][0])
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "signup", nil)
}

func PostSignup(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := r.Form
	switch {
	// if username isn't present or not username field(s) or username is blank
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		err = tem.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Bad Username",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	// if password isn't present or not 2 password field(s) or password is blank
	case f["password"] == nil, len(f["password"]) != 2, f["password"][0] == "":
		err = tem.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Bad Password",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	// if email isn't present or there aren't 1 email field(s) or email is blank
	case f["email"] == nil, len(f["email"]) != 1, f["email"][0] == "":
		err = tem.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Bad Email",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	// if the two passwords don't match
	case f["password"][0] != f["password"][1]:
		err = tem.ExecuteTemplate(w, "signup", map[string]interface{}{
			"Error":"Passwords do not match",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	// otherwise regester user
	default:
		// result := User{}
		var i int
		i, err = muser.Find(bson.M{"username": f["username"][0]}).Count()
		if i < 1 {
			answer, err := bcrypt.GenerateFromPassword(
				[]byte(f["password"][0]), 11)
			if err != nil {
				// bcrypt error
				log.Panic(err)
				return // stop
			}
			err = muser.Insert(&User{
				Id: bson.NewObjectId(),
				Username: f["username"][0],
				Hash: string(answer),
				Email: f["email"][0],
				Joined: time.Now(),
			})
			if err != nil {
				// mgo error
				log.Panic(err)
				io.WriteString(w, "There was an error... Where did it get to?")
				return // stop
			}
			io.WriteString(w, "Thanks for signing up!")
			return // stop
		}else{
			err = tem.ExecuteTemplate(w, "signup", map[string]interface{}{
				"Error":"Username taken",
			})
			if err != nil {
				// html/template error
				log.Panic(err)
			}
			return // stop
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

func GetPeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	user := User{}
	peeves := []Peeve{}
	err = muser.Find(bson.M{"username": c.URLParams["username"]}).One(&user)
	if err != nil {
		if err == mgo.ErrNotFound {
			// user not registered
			err = tem.ExecuteTemplate(w, "error", map[string]interface{}{
				"Number":"404",
				"Body":"Not Found",
			})
			if err != nil {
				log.Panic(err)
			}
			return // stop
		}
		log.Panic(err)
		return // stop
	}
	err = mpeeve.Find(bson.M{"user": user.Id}).All(&peeves)
	if err != nil {
		log.Panic(err)
		return // stop
	}
	if len(peeves) > 0 {
		// if peeves
		err = tem.ExecuteTemplate(w, "user", map[string]interface{}{
			"Peeves": peeves,
		})
		if err != nil {
			log.Panic(err)
			return // stop
		}
	}else{
		// if no peeves
		err = tem.ExecuteTemplate(w, "user", nil)
		if err != nil {
			log.Panic(err)
			return // stop
		}
	}
}

func PostPeeve(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	user := User{}
	peeves := []Peeve{}
	err = muser.Find(bson.M{"username": c.URLParams["username"]}).One(&user)
	if err != nil {
		// user not registered
		if err == mgo.ErrNotFound {
			err = tem.ExecuteTemplate(w, "error", map[string]interface{}{
				"Number":"404",
				"Body":"Not Found",
			})
			if err != err {
				log.Panic(err)
			}
			return // stop
		}
		log.Panic(err)
		return // stop
	}
	err = mpeeve.Find(bson.M{"user": user.Id}).All(&peeves)
	if err != nil {
		log.Panic(err)
		return // stop
	}
	f := r.Form
	switch {
	case f["body"]==nil,len(f["body"]) != 1, f["body"][0] == "":
		err = tem.ExecuteTemplate(w, "user", map[string]interface{}{
			"Peeves": peeves,
			"Error": "Invalid Body",
		})
		if err != nil {
			// html/template error
			log.Panic(err)
		}
		return // stop
	default:
		mpeeve.Insert(&Peeve{
			Id: bson.NewObjectId(),
			Creator: user.Id,
			User: user.Id, // create a peeve == owner
			Body: f["body"][0],
		})
		http.Redirect(w,r,"/"+c.URLParams["username"],302)
	}
}
