package main

import (
	"log"
	"net/http"

	// echo
	"github.com/labstack/echo"
)

func IndexTemplate(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302) // redirect to their page
		return nil // stop
	}
	return c.Render(http.StatusOK, "index", nil) // only serve index.html
}

func SignupTemplate(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302)
		return nil // stop
	}
	return c.Render(http.StatusOK, "signup", nil)
}

func LoginTemplate(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		http.Redirect(w, r, "/"+username, 302)
		return nil // stop
	}
	return c.Render(http.StatusOK, "login", nil)
}

func SettingsTemplate(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string)
	if username == "" {
		// user is not logged in
		http.Redirect(w, r, "/", 302)
		return nil // stop
	}
	if username != c.Param("username") {
		// user is not this user
		http.Redirect(w, r, "/"+username+"/settings", 302)
		return nil // stop
	}
	return c.Render(http.StatusOK, "settings", map[string]interface{}{
		"SessionUsername": username,
	})
}

func Search(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	var username string
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ = session.Values["username"].(string) // convert to string
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// TODO: actually search something instead of just redirect to <user>
	f := r.Form
	switch {
	case f["q"] == nil, len(f["q"]) != 1:
		// if query isn't defined or isn't an array of 1 element
		return c.Render(http.StatusNotFound, "error", map[string]interface{}{
			"Number": "404",
			"Body": "Not Found",
			"SessionUsername": username, // this might be blank
			"Session": session, // this might be blank
		})
		if err != nil {
			log.Panic(err)
			return nil // stop
		}
	case f["q"] != nil, len(f["q"]) == 1:
		users := []user{} // many users can be returned
		go searchUser(f["q"][0], &users, errs)
		switch <-errs {
		case nil:
			break // nil is good
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return nil // stop
		}
		peeves := []peeveAndUser{} // many peeves can be returned
		go searchPeeve(f["q"][0], &peeves, errs)
		switch <-errs {
		case nil:
			break // nil is good
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return nil // stop
		}
		return c.Render(http.StatusOK, "search", map[string]interface{}{
			"Users": users,
			"Peeves": peeves,
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
			return nil // stop
		}
		return nil // stop
	default:
		http.Redirect(w, r, "/"+f["q"][0], 302)
	}
	return nil // stop
}
