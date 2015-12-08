package main

import (
	"log"
	"time"
	"net/http"

	// echo
	"github.com/labstack/echo"

	// rethink
	"gopkg.in/dancannon/gorethink.v1"
)

func GetPeeves(c *echo.Context) error {
	r := c.Request()
	// w := c.Response()
	log.Print(c.Request())
	log.Print(c.Response())
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
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
		// if err != nil {
		// 	log.Panic(err)
		// }
		// return
	default:
		log.Panic(<-errs)
		return nil// stop
	}
	peeves := []peeve{}
	go getPeeves(thisUser.Id, &peeves, errs)
	switch <-errs {
	case nil:
		break
	case gorethink.ErrEmptyResult:
		break // none is okay
	default:
		log.Panic(<-errs)
		return nil // stop
	}
	return c.Render(http.StatusOK, "user", map[string]interface{}{
		"Peeves": peeves,
		"User": thisUser,
		"SessionUsername": username,
		"Session": session,
	})
	// if err != nil {
	// 	log.Panic(err)
	// 	return nil // stop
	// }
}

func CreatePeeve(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.Param("username") { // if user logged isn't this user
		http.Redirect(w, r, "/"+c.Param("username"), 302)
		return nil // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	case f["body"] == nil, len(f["body"]) != 1, f["body"][0] == "":
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid Body",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return nil // stop
	case len(f["body"][0]) > 140:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Peeve Too Long",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return nil // stop
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
			// if err != nil {
			// 	log.Panic(err)
			// 	return nil // stop
			// }
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
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
			log.Panic(<-errs)
		}
		http.Redirect(w, r, "/"+c.Param("username"), 302)
	}
	return nil
}

func DeletePeeve(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username != c.Param("username") { // if user logged isn't this user
		http.Redirect(w, r, "/"+c.Param("username"), 302)
		return nil // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
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
		// if err != nil {
		// 	log.Panic(err)
		// 	return nil // stop
		// }
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
			// if err != nil {
			// 	log.Panic(err)
			// 	return nil // stop
			// }
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
			return nil // stop
		}
		go dropOnePeeve(f["id"][0], user.Id, errs)
		if <-errs != nil {
			log.Panic(<-errs)
			return nil // stop
		}
		http.Redirect(w, r, "/"+c.Param("username"), 302)
	}
	return nil
}

func MeTooPeeve(c *echo.Context) error {
	r := c.Request()
	w := c.Response()
	session, err := redisStore.Get(r, sessionName)
	if err != nil {
		log.Panic(err)
	}
	username, _ := session.Values["username"].(string) // convert to string
	if username == c.Param("username") {
		// don't allow a user to metoo their own peeve
		http.Redirect(w, r, "/"+c.Param("username"), 302)
		return nil // stop
	}
	r.ParseForm() // translate form
	r.ParseMultipartForm(1000000) // translate multipart 1Mb limit
	// don't do anything before we know the form is what we want
	f := r.Form
	switch {
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["id"] == nil, len(f["id"]) != 1, len(f["id"][0]) != 36:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid Id",
			"SessionUsername": username,
			"Session": session,
		})
		// if err != nil {
		// 	log.Panic(err)
		// 	return nil // stop
		// }
	// needs to be length 36 as rethinkdb's ids are len 36
	case f["user"] == nil, len(f["user"]) != 1, len(f["user"][0]) != 36:
		return c.Render(http.StatusOK, "user", map[string]interface{}{
			"Error": "Invalid User",
			"SessionUsername": username,
			"Session": session,
		})
		if err != nil {
			log.Panic(err)
		}
		return nil // stop
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
			// if err != nil {
			// 	log.Panic(err)
			// 	return nil // stop
			// }
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
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
			// if err != nil {
			// 	log.Panic(err)
			// 	return nil // stop
			// }
		default:
			http.Error(w, http.StatusText(500), 500)
			log.Panic(<-errs)
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
			log.Panic(<-errs)
		}
		http.Redirect(w, r, "/"+c.Param("username"), 302)
	}
	return nil
}
