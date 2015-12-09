package main

import (
	"time"
	"net/http"

	// echo
	"github.com/labstack/echo"

	// rethink
	"gopkg.in/dancannon/gorethink.v1"
)

func GetPeeves(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	thisUser := user{}
	go getOneUser(c.Param("username"), &thisUser, errs)
	switch <-errs {
	case nil:
		break
	case gorethink.ErrEmptyResult:
		return c.Render(http.StatusNotFound, "error", map[string]interface{}{
			"Number": "404",
			"Body": "Not Found",
			"SessionUsername": username, // this might be blank
			"Session": session, // this might be blank
		})
	default:
		c.Echo().Logger().Trace(<-errs)
		return nil // stop
	}
	peeves := []peeve{}
	go getPeeves(thisUser.Id, &peeves, errs)
	switch <-errs {
	case nil:
		break
	case gorethink.ErrEmptyResult:
		break // none is okay
	default:
		c.Echo().Logger().Trace(<-errs)
		return nil // stop
	}
	return c.Render(http.StatusOK, "user", map[string]interface{}{
		"Peeves": peeves,
		"User": thisUser,
		"SessionUsername": username,
		"Session": session,
	})
}

func CreatePeeve(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.Param("username") { // if user logged isn't this user
		return c.Redirect(302, "/"+c.Param("username"))
	}
	c.Request().ParseForm() // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := c.Request().Form
	switch {
	case f["body"] == nil, len(f["body"]) != 1, f["body"][0] == "":
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid Body",
			"SessionUsername": username,
			"Session": session,
		})
	case len(f["body"][0]) > 140:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Peeve Too Long",
			"SessionUsername": username,
			"Session": session,
		})
	default:
		user := user{}
		go getOneUser(c.Param("username"), &user, errs)
		switch <-errs {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			return c.Render(http.StatusNotFound, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
				"Session": session,
			})
		default:
			http.Error(c.Response(), http.StatusText(500), 500)
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
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
			c.Echo().Logger().Trace(<-errs)
		}
		return c.Redirect(302, "/"+c.Param("username"))
	}
	return nil
}

func DeletePeeve(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.Param("username") { // if user logged isn't this user
		return c.Redirect(302, "/"+c.Param("username"))
	}
	c.Request().ParseForm() // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := c.Request().Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid Id",
			// "Peeves": peeves,
			// "User": user,
			"SessionUsername": username,
			"Session": session,
		})
	default:
		user := user{}
		go getOneUser(username, &user, errs)
		switch <-errs {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			return c.Render(http.StatusNotFound, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
				"Session": session,
			})
		default:
			http.Error(c.Response(), http.StatusText(500), 500)
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
		}
		go dropOnePeeve(f["id"][0], user.Id, errs)
		if <-errs != nil {
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
		}
		return c.Redirect(302, "/"+c.Param("username"))
	}
	return nil
}

func MeTooPeeve(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Trace(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username == c.Param("username") {
		// don't allow a user to metoo their own peeve
		return c.Redirect(302, "/"+c.Param("username"))
	}
	c.Request().ParseForm() // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := c.Request().Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid Id",
			"SessionUsername": username,
			"Session": session,
		})
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["user"] == nil, len(f["user"]) != 1, len(f["user"][0]) != 36:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid User",
			"SessionUsername": username,
			"Session": session,
		})
	default:
		user := user{}
		go getOneUser(username, &user, errs)
		switch <-errs {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			return c.Render(http.StatusNotFound, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
				"Session": session,
			})
		default:
			http.Error(c.Response(), http.StatusText(500), 500)
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
		}
		metoopeeve := peeve{}
		// don't assume input is valid
		go getOnePeeve(f["id"][0], f["user"][0], &metoopeeve, errs)
		switch <-errs {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			return c.Render(http.StatusNotFound, "error", map[string]interface{}{
				"Number": "404",
				"Body": "Not Found",
				"SessionUsername": username,
				"Session": session,
			})
		default:
			http.Error(c.Response(), http.StatusText(500), 500)
			c.Echo().Logger().Trace(<-errs)
			return nil // stop
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
			c.Echo().Logger().Trace(<-errs)
		}
		return c.Redirect(302, "/"+c.Param("username"))
	}
	return nil
}
