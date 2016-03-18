package main

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	// rethink
	r "gopkg.in/dancannon/gorethink.v1"

	// redis
	redis "gopkg.in/boj/redistore.v1"
)

func getRethinkSession(sessionChan chan *r.Session) {
	var rethinkurl []string
	var rethinkauth string
	var rethinkdb string
	var err error
	switch {
	// simple
	case os.Getenv("RETHINK") != "":
		rethinkurl = []string{os.Getenv("RETHINK")}
		if len(strings.Split(rethinkurl[0], ";")) > 1 {
			rethinkurl = strings.Split(rethinkurl[0], ";")
		}
		if len(strings.Split(rethinkurl[0], ",")) > 1 {
			rethinkurl = strings.Split(rethinkurl[0], ",")
		}
	// docker
	case os.Getenv("RETHINK_PORT_28015_TCP_ADDR") != "" &&
		os.Getenv("RETHINK_PORT_28015_TCP_PORT") != "":
		rethinkurl = []string{os.Getenv("RETHINK_PORT_28015_TCP_ADDR") + ":" +
			os.Getenv("RETHINK_PORT_28015_TCP_PORT")}
	// default fail
	default:
		log.Fatal("RETHINK Env Undefined")
	}
	if os.Getenv("RETHINK_DB") != "" {
		rethinkdb = os.Getenv("RETHINK_DB")
	} else {
		rethinkdb = "whatannoysme"
	}
	if os.Getenv("RETHINK_AUTH") != "" {
		rethinkauth = os.Getenv("RETHINK_AUTH")
	}
	session, err := r.Connect(r.ConnectOpts{
		Addresses:     rethinkurl,
		AuthKey:       rethinkauth,
		Database:      rethinkdb,
		MaxIdle:       10,
		MaxOpen:       40,
		DiscoverHosts: true,
	})
	session.Use(rethinkdb)
	if err != nil {
		log.Fatal(err)
	}
	sessionChan <- session
}

func getRediStore(redisChan chan *redis.RediStore) {
	var redisHostPort string = ":6379"
	var redisPassword string = ""
	redisClients, err := strconv.Atoi(os.Getenv("REDIS_CLIENTS"))
	if err != nil {
		log.Print(err)
		// assume undefined
		redisClients = 2
	}
	// auth?
	if os.Getenv("REDIS_PASS") != "" {
		redisPassword = os.Getenv("REDIS_PASS")
	}
	switch {
	// simple
	case os.Getenv("REDIS") != "":
		redisURL, err := url.Parse(os.Getenv("REDIS"))
		if err != nil {
			log.Print(err)
		}
		if redisURL.User != nil {
			redisPassword, _ = redisURL.User.Password()
		}
		redisHostPort = redisURL.Host
		if redisHostPortArray := strings.Split(redisURL.Host, ":"); len(redisHostPortArray) < 2 {
			// if the host can't be split by : then append default redis port
			redisHostPort += ":6379"
		}
	// docker
	case os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_PORT") != "":
		redisHostPort = os.Getenv("REDIS_PORT_6379_TCP_ADDR") + ":" +
			os.Getenv("REDIS_PORT_6379_TCP_PORT")
	// default fail
	default:
		log.Fatal("REDIS Env Undefined")
	}
	redisStore, err := redis.NewRediStore(redisClients,
		"tcp", redisHostPort, redisPassword, []byte(os.Getenv("KEY")))
	if err != nil {
		log.Fatal(err)
	}
	redisChan <- redisStore
}

// create

func createUser(user interface{}, done chan error) {
	_, err := r.Table("users").Insert(user).RunWrite(rethinkSession)
	if err != nil {
		log.Print("createUser", err)
	}
	done <- err
}

func createPeeve(peeve interface{}, done chan error) {
	_, err := r.Table("peeves").Insert(peeve).RunWrite(rethinkSession)
	if err != nil {
		log.Print("createPeeve", err)
	}
	done <- err
}

// get

func getUsers(username string, users interface{}, done chan error) {
	cursor, err := r.Table("users").Filter(map[string]interface{}{
		"username": username,
	}).Run(rethinkSession)
	if err != nil {
		log.Print("getUsers", err)
		done <- err
		return // stop
	}
	defer cursor.Close()
	done <- cursor.All(users)
}

func getPeeves(userId string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(map[string]interface{}{
		"user": userId,
	}).OrderBy(
		"timestamp",
	).Run(rethinkSession)
	if err != nil {
		log.Print("getPeeves", err)
		done <- err
		return // stop
	}
	defer cursor.Close()
	done <- cursor.All(peeves)
}

// get one

func getOneUser(username string, user interface{}, done chan error) {
	cursor, err := r.Table("users").Filter(map[string]interface{}{
		"username": username,
	}).Run(rethinkSession)
	if err != nil {
		log.Print("getOneUser", err)
		done <- err
		return
	}
	defer cursor.Close()
	done <- cursor.One(user)
}

func getOnePeeve(peeveId string, userId string, peeve interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(map[string]interface{}{
		"id":   peeveId,
		"user": userId,
	}).Run(rethinkSession)
	if err != nil {
		log.Print("getOnePeeve", err)
		done <- err
		return
	}
	defer cursor.Close()
	done <- cursor.One(peeve)
}

// get count

func getCountUsername(username string, count interface{}, done chan error) {
	cursor, err := r.Table("users").Filter(map[string]interface{}{
		"username": username,
	}).Count().Run(rethinkSession)
	if err != nil {
		log.Print("getCountUsername ", err)
		done <- err
		return
	}
	defer cursor.Close()
	done <- cursor.One(count)
}

// search

func searchUser(search string, users interface{}, done chan error) {
	cursor, err := r.Table("users").
		Filter(r.Row.Field("username").Match(search)).
		Run(rethinkSession)
	if err != nil {
		log.Print("searchUser ", err)
		done <- err
		return
	}
	defer cursor.Close()
	done <- cursor.All(users)
}

func searchPeeve(query string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").
		Filter(r.Row.Field("body").Match(query)).
		Filter(r.Row.Field("user").Eq(r.Row.Field("root"))).
		EqJoin("user", r.Table("users")).
		Zip().
		Run(rethinkSession)
	if err != nil {
		log.Print("searchPeeve", err)
		done <- err
		return
	}
	defer cursor.Close()
	done <- cursor.All(peeves)
}

func searchPeeveField(query string, field string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").
		Filter(r.Row.Field(field).Match(query)).
		Run(rethinkSession)
	if err != nil {
		log.Print("searchPeeveField", err)
		done <- err
		return
	}
	defer cursor.Close()
	done <- cursor.All(peeves)
}

// drop ones

func dropOneUser(userId string, done chan error) {
	log.Fatal("dont run this")
	_, err := r.Table("users").Get(userId).Limit(1).Delete().RunWrite(rethinkSession)
	done <- err
}

func dropOnePeeve(peeveId string, userId string, done chan error) {
	_, err := r.Table("peeves").
		Filter(map[string]interface{}{
			"id":   peeveId,
			"user": userId,
		}).Limit(1).Delete().
		RunWrite(rethinkSession)
	done <- err
}
