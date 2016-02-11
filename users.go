package main

import (
	"net/http"
	"strings"
	"time"

	// golang.org/x/crypto
	"golang.org/x/crypto/bcrypt"

	// echo
	"github.com/labstack/echo"

	// rethink
	"gopkg.in/dancannon/gorethink.v1"
)

func CreateUser(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Debug(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		return c.Redirect(302, "/"+username)
	}
	c.Request().ParseForm()                 // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request().Form
	switch {
	// if username isn't present or there aren't username field(s) or blank
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Bad Username",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	// if password isn't present or there aren't 2 password field(s) or blank
	case f["password"] == nil, len(f["password"]) != 2, f["password"][0] == "":
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Bad Password",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	// if email isn't present or there aren't 1 email field(s) or email is blank
	case f["email"] == nil, len(f["email"]) != 1, f["email"][0] == "":
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Bad Email",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	// max username length 13
	case len(f["username"][0]) > 13:
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Username too long",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	// if the two passwords don't match
	case f["password"][0] != f["password"][1]:
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Passwords do not match",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	}
	// otherwise regester user
	if len(strings.Fields(f["username"][0])) > 1 {
		// username has \t \n or space
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Username Contains Invalid Characters",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	}
	// force all usernames to be lowercase
	f["username"][0] = strings.ToLower(f["username"][0])
	var i int
	go getCountUsername(f["username"][0], &i, errs)
	if <-errs != nil {
		c.Echo().Logger().Debug(<-errs)
	}
	if i > 1 {
		return c.Render(http.StatusOK, "signup", map[string]interface{}{
			"Error": "Username taken",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	}
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(f["password"][0]), bcryptStrength)
	if err != nil {
		c.Echo().Logger().Debug(err)
		return nil // stop
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
		http.Error(c.Response(), http.StatusText(500), 500)
		c.Echo().Logger().Debug(<-errs)
		return nil // stop
	}
	// session.Values["user"] = user
	session.Values["username"] = newuser.Username
	session.Values["hash"] = string(hash)
	if err = session.Save(c.Request(), c.Response()); err != nil {
		c.Echo().Logger().Debug("Error saving session: %v", err)
	}
	return c.Redirect(302, "/"+f["username"][0])
}

func Login(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Debug(err)
	}
	username, _ := session.Values["username"].(string)
	if username != "" {
		return c.Redirect(302, "/"+username)
	}
	c.Request().ParseForm()                 // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request().Form
	switch {
	case f["username"] == nil, len(f["username"]) != 1, f["username"][0] == "":
		return c.Render(http.StatusOK, "login", map[string]interface{}{
			"Error": "Invalid Username",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	case f["password"] == nil, len(f["password"]) != 1, f["password"][0] == "":
		return c.Render(http.StatusOK, "login", map[string]interface{}{
			"Error": "Invalid Password",
		})
		if err != nil {
			c.Echo().Logger().Debug(err)
		}
		return nil // stop
	default:
		f["username"][0] = strings.ToLower(f["username"][0]) // assure one user per username
		user := user{}
		go getOneUser(f["username"][0], &user, errs)
		switch <-errs {
		case nil:
			break
		case gorethink.ErrEmptyResult:
			return c.Render(http.StatusOK, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			if err != nil {
				c.Echo().Logger().Debug(err)
			}
			return nil // stop
		default:
			c.Echo().Logger().Debug(<-errs)
			return nil // stop
		}
		// user found
		err = bcrypt.CompareHashAndPassword([]byte(user.Hash),
			[]byte(f["password"][0]))
		switch err {
		case nil:
			break
		case bcrypt.ErrMismatchedHashAndPassword:
			// incorrect password
			return c.Render(http.StatusOK, "login", map[string]interface{}{
				"Error": "Invalid Username or Password",
			})
			if err != nil {
				c.Echo().Logger().Debug(err)
			}
			return nil // stop
		default:
			c.Echo().Logger().Debug(err)
			return nil // stop
		}
		// correct password
		// session.Values["user"] = user
		session.Values["username"] = user.Username
		session.Values["hash"] = user.Hash
		if err = session.Save(c.Request(), c.Response()); err != nil {
			c.Echo().Logger().Debug("Error saving session: %v", err)
		}
		return c.Redirect(302, "/"+f["username"][0])
	}
	return nil // stop
}

func Logout(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Debug(err)
	}
	session.Options.MaxAge = -1
	if err = session.Save(c.Request(), c.Response()); err != nil {
		c.Echo().Logger().Debug("Error saving session: %v", err)
	}
	return c.Redirect(302, "/")
}

func Settings(c *echo.Context) error {
	session, err := redisStore.Get(c.Request(), sessionName)
	if err != nil {
		c.Echo().Logger().Debug(err)
	}
	username, _ := session.Values["username"].(string)
	if username != c.Param("username") {
		// user is not this user
		return c.Redirect(302, "/"+c.Param("username")+"/settings")
	}
	thisuser := user{}
	go getOneUser(c.Param("username"), &thisuser, errs)
	if <-errs != nil {
		c.Echo().Logger().Debug(<-errs)
		return nil // stop
	}
	c.Request().ParseForm()                 // translate form
	c.Request().ParseMultipartForm(1000000) // translate multipart 1Mb limit
	f := c.Request().Form
	if len(f) > 0 {
		if len(f["first"]) == 1 {
			thisuser.setFirstName(f["first"][0])
		}
		if len(f["last"]) == 1 {
			thisuser.setLastName(f["last"][0])
		}
		if len(f["password"]) == 2 && f["password"][0] == f["password"][1] {
			thisuser.setPassword(f["password"][0])
		}
	} else {
		http.Error(c.Response(), http.StatusText(500), 500)
		return nil // stop
	}
	return c.Redirect(302, "/"+c.Param("username")+"/settings")
}
