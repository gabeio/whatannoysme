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
	case os.Getenv("REDIS_PASS") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_PORT") != "":
		redisHostPort = os.Getenv("REDIS_PORT_6379_TCP_ADDR") + ":" +
			os.Getenv("REDIS_PORT_6379_TCP_PORT")
		redisPassword = os.Getenv("REDIS_PASS")
	case os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_PORT") != "":
		redisHostPort = os.Getenv("REDIS_PORT_6379_TCP_ADDR") + ":" +
			os.Getenv("REDIS_PORT_6379_TCP_PORT")
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
	cursor, err := r.Table("users").Filter(func (row r.Term) r.Term {
		return row.Field("username").Match(search)
	}).Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(users)
}

func searchPeeve(search string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(func (row r.Term) r.Term {
		return row.Field("body").Match(search)
	}).Filter(func (row r.Term) r.Term {
		return row.Field("user").Eq(row.Field("root"))
	}).InnerJoin(r.Table("users"), func (peeveRow r.Term, userRow r.Term) r.Term {
		return peeveRow.Field("user").Eq(userRow.Field("id"))
	}).Zip().Run(rethinkSession)
	defer cursor.Close()
	if err != nil {
		log.Panic(err)
	}
	done<- cursor.All(peeves)
}

func searchPeeveField(search string, field string, peeves interface{}, done chan error) {
	cursor, err := r.Table("peeves").Filter(func (row r.Term) r.Term {
		return row.Field(field).Match(search)
	}).Run(rethinkSession)
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
	}).Delete().RunWrite(rethinkSession)
	done<- err
}
