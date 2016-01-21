package main

import (
	"os"
	"log"
	"strconv"
	"strings"
	"net/url"

	// rethink
	r "gopkg.in/dancannon/gorethink.v1"

	// redis
	redis "gopkg.in/boj/redistore.v1"
)

func getRethinkSession(sessionChan chan *r.Session) {
	var rethinkurl string
	var err error
	switch {
	// simple
	case os.Getenv("RETHINK") != "":
		rethinkurl = os.Getenv("RETHINK")
	// docker
	case os.Getenv("RETHINK_PORT_28015_TCP_ADDR") != "" &&
		os.Getenv("RETHINK_PORT_28015_TCP_PORT") != "":
		rethinkurl = os.Getenv("RETHINK_PORT_28015_TCP_ADDR") + ":" +
			os.Getenv("RETHINK_PORT_28015_TCP_PORT")
	// default fail
	default:
		log.Fatal("RETHINK Env Undefined")
	}
	session, err := r.Connect(r.ConnectOpts{
		Address: rethinkurl,
		Database: "whatannoysme",
	    MaxIdle: 1,
	    MaxOpen: 10,
		// DiscoverHosts: true,
	})
	session.Use(os.Getenv("RETHINK_DB"))
	if err != nil {
		log.Fatal(err)
	}
	sessionChan<- session
}

func getRediStore(redisChan chan *redis.RediStore) {
	// var redisStore *redistore.RediStore = new(redistore.RediStore)
	var redisHostPort string = ":6379"
	var redisPassword string = ""
	redisClients, err := strconv.Atoi(os.Getenv("REDIS_CLIENTS"))
	if err != nil {
		log.Panic(err)
		// assume undefined
		redisClients = 2
	}
	switch {
	// simple
	case os.Getenv("REDIS") != "":
		redisURL, err := url.Parse(os.Getenv("REDIS"))
		if err != nil {
			log.Panic(err)
		}
		if redisURL.User != nil {
			redisPassword, _ = redisURL.User.Password()
		}
		redisHostPort = redisURL.Host
		if redisHostPortArray := strings.Split(redisURL.Host, ":");
			len(redisHostPortArray) < 2 {
			// if the host can't be split by : then append default redis port
			redisHostPort += ":6379"
		}
	// docker with auth
	case os.Getenv("REDIS_PASS") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_PORT") != "":
		redisHostPort = os.Getenv("REDIS_PORT_6379_TCP_ADDR") + ":" +
			os.Getenv("REDIS_PORT_6379_TCP_PORT")
		redisPassword = os.Getenv("REDIS_PASS")
	// docker without auth
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
	redisChan<- redisStore
}

// create

func createUser(user interface{}, done chan error) {
	_, err := r.Table("users").Insert(user).RunWrite(rethinkSession)
	done<- err
}

func createPeeve(peeve interface{}, done chan error) {
	_, err := r.Table("peeves").Insert(peeve).RunWrite(rethinkSession)
	done<- err
}

// get

func getUsers(username string, users interface{}, done chan error) {
	cursor, err := r.Table("users").Filter(map[string]interface{}{
		"username": username,
	}).Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(users)
}

func getPeeves(userId string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(map[string]interface{}{
		"user": userId,
	}).OrderBy(
		"timestamp",
	).Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(peeves)
}

// get one

func getOneUser(username string, user interface{}, done chan error) {
	cursor, err := r.Table("users").Filter(map[string]interface{}{
		"username": username,
	}).Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.One(user)
}

func getOnePeeve(peeveId string, userId string, peeve interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(map[string]interface{}{
		"id": peeveId,
		"user": userId,
	}).Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.One(peeve)
}

// get count

func getCountUsername(username string, count interface{}, done chan error) {
	cursor, err := r.DB("whatannoysme").Table("users").Filter(map[string]interface{}{
		"username": username,
	}).Count().Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.One(count)
}

// search

func searchUser(search string, users interface{}, done chan error) {
	cursor, err := r.Table("users").
		Filter(r.Row.Field("username").
		Match(search)).
		Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(users)
}

func searchPeeve(query string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").
		Filter(r.Row.Field("body").Match(query)).
		Filter(r.Row.Field("user").Eq(r.Row.Field("root"))).
		EqJoin("user", r.Table("users")).
		Zip().
		Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(peeves)
}

func searchPeeveField(query string, field string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(r.Row.Field(field).Match(query)).Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(peeves)
}

// drop one

func dropOneUser(userId string, done chan error) {
	log.Fatal("dont run this")
	_, err := r.Table("users").Get(userId).Delete().RunWrite(rethinkSession)
	done<- err
}

func dropOnePeeve(peeveId string, userId string, done chan error) {
	_, err := r.Table("peeves").Filter(map[string]interface{}{
		"id": peeveId,
		"user": userId,
	}).Limit(1).Delete().RunWrite(rethinkSession)
	done<- err
}
