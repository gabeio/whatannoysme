package main

import (
	"log"
	"net/http"
	"time"

	// gin
	"github.com/gin-gonic/gin"

	// rethink
	"gopkg.in/dancannon/gorethink.v1"
)

func GetPeeves(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	errs := make(chan error)
	defer close(errs)
	username, _ := session.Values["username"].(string) // convert to string
	thisUser := user{}
	go getOneUser(c.Param("username"), &thisUser, errs)
	err = <-errs
	switch err {
	case nil:
		break
	case gorethink.ErrEmptyResult:
		c.HTML(http.StatusNotFound, "error", map[string]interface{}{
			"Number":          "404",
			"Body":            "Not Found",
			"SessionUsername": username, // this might be blank
			"Session":         session,  // this might be blank
		})
		return // stop
	default:
		log.Print(err)
		return // stop
	}
	peeves := []peeve{}
	go getPeeves(thisUser.Id, &peeves, errs)
	err = <-errs
	switch err {
	case nil:
		break
	case gorethink.ErrEmptyResult:
		break // none is okay
	default:
		log.Print(err)
		return // stop
	}
	c.HTML(http.StatusOK, "user", map[string]interface{}{
		"Peeves":          peeves,
		"User":            thisUser,
		"SessionUsername": username,
		"Session":         session,
	})
}

func CreatePeeve(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	errs := make(chan error)
	defer close(errs)
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.Param("username") {               // if user logged isn't this user
		c.Redirect(302, "/u/"+c.Param("username"))
		return
	}
	c.Request.ParseForm()                 // translate form
	c.Request.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := c.Request.Form
	switch {
	case f["body"] == nil, len(f["body"]) != 1, f["body"][0] == "":
		c.HTML(http.StatusOK, "user", map[string]interface{}{
			"Error":           "Invalid Body",
			"SessionUsername": username,
			"Session":         session,
		})
		return
	case len(f["body"][0]) > 140:
		c.HTML(http.StatusOK, "user", map[string]interface{}{
			"Error":           "Peeve Too Long",
			"SessionUsername": username,
			"Session":         session,
		})
		return
	default:
		user := user{}
		go getOneUser(c.Param("username"), &user, errs)
		err = <-errs
		switch err {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			c.HTML(http.StatusNotFound, "error", map[string]interface{}{
				"Number":          "404",
				"Body":            "Not Found",
				"SessionUsername": username,
				"Session":         session,
			})
			return
		default:
			http.Error(c.Writer, http.StatusText(500), 500)
			log.Print(err)
			return // stop
		}
		go createPeeve(&peeve{
			// Id: bson.NewObjectId(),
			Root: user.Id,
			// as this is the root no parent
			UserId:    user.Id, // create a peeve == owner
			Body:      f["body"][0],
			Timestamp: time.Now(),
		}, errs)
		if err = <-errs; err != nil {
			log.Print(err)
		}
		c.Redirect(302, "/u/"+c.Param("username"))
		return
	}
	return
}

func DeletePeeve(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	errs := make(chan error)
	defer close(errs)
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.Param("username") {               // if user logged isn't this user
		c.Redirect(302, "/u/"+c.Param("username"))
		return
	}
	c.Request.ParseForm()                 // translate form
	c.Request.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := c.Request.Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		c.HTML(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid Id",
			// "Peeves": peeves,
			// "User": user,
			"SessionUsername": username,
			"Session":         session,
		})
		return
	default:
		user := user{}
		go getOneUser(username, &user, errs)
		err = <-errs
		switch err {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			c.HTML(http.StatusNotFound, "error", map[string]interface{}{
				"Number":          "404",
				"Body":            "Not Found",
				"SessionUsername": username,
				"Session":         session,
			})
			return
		default:
			http.Error(c.Writer, http.StatusText(500), 500)
			log.Print(err)
			return // stop
		}
		go dropOnePeeve(f["id"][0], user.Id, errs)
		if err = <-errs; err != nil {
			log.Print(err)
			return // stop
		}
		c.Redirect(302, "/u/"+c.Param("username"))
		return
	}
	return
}

func MeTooPeeve(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	errs := make(chan error)
	defer close(errs)
	username, _ := session.Values["username"].(string) // convert to string
	if username == c.Param("username") {
		// don't allow a user to metoo their own peeve
		c.Redirect(302, "/")
		return
	}
	c.Request.ParseForm()                 // translate form
	c.Request.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := c.Request.Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		c.HTML(http.StatusOK, "user", map[string]interface{}{
			"Error":           "Invalid Id",
			"SessionUsername": username,
			"Session":         session,
		})
		return
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["user"] == nil, len(f["user"]) != 1, len(f["user"][0]) != 36:
		c.HTML(http.StatusOK, "user", map[string]interface{}{
			"Error":           "Invalid User",
			"SessionUsername": username,
			"Session":         session,
		})
		return
	default:
		user := user{}
		go getOneUser(username, &user, errs)
		err = <-errs
		switch err {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			c.HTML(http.StatusNotFound, "error", map[string]interface{}{
				"Number":          "404",
				"Body":            "Not Found",
				"SessionUsername": username,
				"Session":         session,
			})
			return
		default:
			http.Error(c.Writer, http.StatusText(500), 500)
			log.Print(err)
			return // stop
		}
		metoopeeve := peeve{}
		// don't assume input is valid
		go getOnePeeve(f["id"][0], f["user"][0], &metoopeeve, errs)
		err = <-errs
		switch err {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			c.HTML(http.StatusNotFound, "error", map[string]interface{}{
				"Number":          "404",
				"Body":            "Not Found",
				"SessionUsername": username,
				"Session":         session,
			})
			return
		default:
			http.Error(c.Writer, http.StatusText(500), 500)
			log.Print(err)
			return // stop
		}
		peevey := &peeve{
			// Id: bson.NewObjectId(), // create new id
			Root:      metoopeeve.Root,   // peeve origin
			Parent:    metoopeeve.UserId, // who I got it from
			UserId:    user.Id,           // who owns this peeve
			Body:      metoopeeve.Body,   // original body
			Timestamp: time.Now(),        // when I reposted it
		}
		go createPeeve(peevey, errs)
		if err = <-errs; err != nil {
			log.Print(err)
		}
		c.Redirect(302, "/")
		return
	}
	return
}
