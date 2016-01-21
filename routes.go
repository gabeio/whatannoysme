package main

import (
	"net/http"

	// echo
	"github.com/labstack/echo"
)

func IndexTemplate(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		return c.Redirect(302, "/"+username) // redirect to their page
	}
	return c.Render(http.StatusOK, "index", nil) // only serve index.html
}

func SignupTemplate(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		return c.Redirect(302, "/"+username)
	}
	return c.Render(http.StatusOK, "signup", nil)
}

func LoginTemplate(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		return c.Redirect(302, "/"+username)
	}
	return c.Render(http.StatusOK, "login", nil)
}

func SettingsTemplate(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string)
	if username == "" {
		// user is not logged in
		return c.Redirect(302, "/")
	}
	if username != c.Param("username") {
		// user is not this user
		return c.Redirect(302, "/"+username+"/settings")
	}
	return c.Render(http.StatusOK, "settings", map[string]interface{}{
		"SessionUsername": username,
	})
}

func Search(c *echo.Context) error {
	var username string
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ = session.Values["username"].(string) // convert to string
	c.Request().ParseForm() // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request().Form
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
			c.Echo().Logger().Trace(err)
			return nil // stop
		}
	case f["q"] != nil, len(f["q"]) == 1:
		users := []user{} // many users can be returned
		go searchUser(f["q"][0], &users, errs)
		switch <-errs {
		case nil:
			break // nil is good
		default:
			http.Error(c.Response(), http.StatusText(500), 500)
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
		}
		peeves := []peeveAndUser{} // many peeves can be returned
		go searchPeeve(f["q"][0], &peeves, errs)
		switch <-errs {
		case nil:
			break // nil is good
		default:
			http.Error(c.Response(), http.StatusText(500), 500)
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
		}
		return c.Render(http.StatusOK, "search", map[string]interface{}{
			"Users": users,
			"Peeves": peeves,
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			c.Echo().Logger().Trace(err)
			return nil // stop
		}
		return nil // stop
	default:
		return c.Redirect(302, "/"+f["q"][0])
	}
	return nil // stop
}
