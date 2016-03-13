package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	// golang.org/x/crypto
	"golang.org/x/crypto/bcrypt"

	// gin
	"github.com/gin-gonic/gin"

	// rethink
	"gopkg.in/dancannon/gorethink.v1"
)

func CreateUser(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		c.Redirect(302, "/"+username)
	}
	c.Request.ParseForm()                 // translate form
	c.Request.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request.Form
	switch {
	// if username isn't present or there aren't username field(s) or blank
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Bad Username",
		})
		return // stop
	// if password isn't present or there aren't 2 password field(s) or blank
	case f["password"] == nil, len(f["password"]) != 2, f["password"][0] == "":
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Bad Password",
		})
		return // stop
	// if email isn't present or there aren't 1 email field(s) or email is blank
	case f["email"] == nil, len(f["email"]) != 1, f["email"][0] == "":
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Bad Email",
		})
		return // stop
	// max username length 13
	case len(f["username"][0]) > 13:
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Username too long",
		})
		return // stop
	// if the two passwords don't match
	case f["password"][0] != f["password"][1]:
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Passwords do not match",
		})
		return // stop
	}
	// otherwise regester user
	if len(strings.Fields(f["username"][0])) > 1 {
		// username has \t \n or space
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Username Contains Invalid Characters",
		})
		return // stop
	}
	// force all usernames to be lowercase
	f["username"][0] = strings.ToLower(f["username"][0])
	var i int
	go getCountUsername(f["username"][0], &i, errs)
	if <-errs != nil {
		log.Print(<-errs)
	}
	if i > 1 {
		c.HTML(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Username taken",
		})
		return // stop
	}
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(f["password"][0]), bcryptStrength)
	if err != nil {
		log.Print(err)
		return // stop
	}
	newuser := &user{
		// Id: bson.NewObjectId(),
		Username: f["username"][0],
		Hash:     string(hash),
		Email:    f["email"][0],
		Joined:   time.Now(),
	}
	go createUser(newuser, errs)
	if <-errs != nil {
		http.Error(c.Writer, http.StatusText(500), 500)
		log.Print(<-errs)
		return // stop
	}
	// session.Values["user"] = user
	session.Values["username"] = newuser.Username
	session.Values["hash"] = string(hash)
	if err = session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: %v", err)
	}
	c.Redirect(302, "/"+f["username"][0])
}

func Login(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print("Login.session",err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		c.Redirect(302, "/"+username)
	}
	c.Request.ParseForm()                 // translate form
	c.Request.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request.Form
	switch {
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		c.HTML(http.StatusOK, "login", map[string]interface{}{
			"Error": "Invalid Username",
		})
		return // stop
	case f["password"] == nil, len(f["password"]) != 1, f["password"][0] == "":
		c.HTML(http.StatusOK, "login", map[string]interface{}{
			"Error": "Invalid Password",
		})
		return // stop
	default:
		f["username"][0] = strings.ToLower(f["username"][0]) // assure one user per username
		user := user{}
		go getOneUser(f["username"][0], &user, errs)
		switch <-errs {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			c.HTML(http.StatusOK, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			return // stop
		default:
			log.Print("Login.default>default[0]",<-errs)
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
			c.HTML(http.StatusOK, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			return // stop
		default:
			log.Print("Login.default.default[1]",err)
			return // stop
		}
		// correct password
		// session.Values["user"] = user
		session.Values["username"] = user.Username
		session.Values["hash"] = user.Hash
		if err = session.Save(c.Request, c.Writer); err != nil {
			log.Print("Error saving session: %v", err)
		}
		c.Redirect(302, "/"+f["username"][0])
	}
	return // stop
}

func Logout(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print("Logout",err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(c.Request, c.Writer); err != nil {
		log.Print("Error saving session: %v", err)
	}
	c.Redirect(302, "/")
}

func Settings(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print("Settings",err)
	}
	username, _ := session.Values["username"].(string)
	if username != c.Param("username") {
		// user is not this user
		c.Redirect(302, "/"+c.Param("username")+"/settings")
	}
	thisuser := user{}
	go getOneUser(c.Param("username"), &thisuser, errs)
	if <-errs != nil {
		log.Print("getOneUser",<-errs)
		return // stop
	}
	c.Request.ParseForm()                 // translate form
	c.Request.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request.Form
	if len(f) > 0 {
		if len(f["first"]) == 1 {
			err = thisuser.setFirstName(f["first"][0])
			if err != nil {
				log.Print("user.Settings",err)
			}
		}
		if len(f["last"]) == 1 {
			err = thisuser.setLastName(f["last"][0])
			if err != nil {
				log.Print("user.Settings",err)
			}
		}
		if len(f["password"]) == 2 && f["password"][0] == f["password"][1] {
			err = thisuser.setPassword(f["password"][0])
			if err != nil {
				log.Print("user.Settings",err)
			}
		}
	} else {
		http.Error(c.Writer, http.StatusText(500), 500)
		return // stop
	}
	c.Redirect(302, "/"+c.Param("username")+"/settings")
}
