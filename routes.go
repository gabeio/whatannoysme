package main

import (
	"log"
	"net/http"

	// gin
	"github.com/gin-gonic/gin"
)

func IndexTemplate(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		c.Redirect(302, "/u/"+username) // redirect to their page
		return
	}
	c.HTML(http.StatusOK, "index", nil) // only serve index.html
}

func SignupTemplate(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		c.Redirect(302, "/u/"+username)
		return
	}
	c.HTML(http.StatusOK, "signup", nil)
}

func LoginTemplate(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		c.Redirect(302, "/u/"+username)
		return
	}
	c.HTML(http.StatusOK, "login", nil)
}

func SettingsTemplate(c *gin.Context) {
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	username, _ := session.Values["username"].(string)
	if username == "" {
		// user is not logged in
		c.Redirect(302, "/")
		return
	}
	c.HTML(http.StatusOK, "settings", map[string]interface{}{
		"SessionUsername": username,
	})
}

func Search(c *gin.Context) {
	var username string
	session, err := redisStore.Get(c.Request, sessionName)
	if err != nil {
		log.Print(err)
	}
	username, _ = session.Values["username"].(string) // convert to string
	c.Request.ParseForm()                             // translate form
	c.Request.ParseMultipartForm(1000000)             // translate multipart 1Mb limit
	f := c.Request.Form
	switch {
	case f["q"] == nil, len(f["q"]) != 1:
		// if query isn't defined or isn't an array of 1 element
		c.HTML(http.StatusNotFound, "error", map[string]interface{}{
			"Number":          "404",
			"Body":            "Not Found",
			"SessionUsername": username, // this might be blank
			"Session":         session,  // this might be blank
		})
		return
	case f["q"] != nil, len(f["q"]) == 1:
		users := []userModel{} // many users can be returned
		errs := make(chan error)
		defer close(errs)
		go searchUser(f["q"][0], &users, errs)
		err = <-errs
		switch err {
		case nil:
			break // nil is good
		default:
			http.Error(c.Writer, http.StatusText(500), 500)
			log.Print(err)
			return // stop
		}
		peeves := []peeveAndUserModel{} // many peeves can be returned
		go searchPeeve(f["q"][0], &peeves, errs)
		err = <-errs
		switch err {
		case nil:
			break // nil is good
		default:
			http.Error(c.Writer, http.StatusText(500), 500)
			log.Print(err)
			return // stop
		}
		c.HTML(http.StatusOK, "search", map[string]interface{}{
			"Users":           users,
			"Peeves":          peeves,
			"SessionUsername": username,
			"Session":         session,
		})
		return // stop
	default:
		c.Redirect(302, "/"+f["q"][0])
	}
	return // stop
}
