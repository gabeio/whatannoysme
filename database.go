package main

import (
	"os"
	"log"
	"strconv"
	"strings"
	"net/url"

	// mongo
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// redis
	// "github.com/garyburd/redigo/redis"
	"gopkg.in/boj/redistore.v1"
)

func getMgoSession() *mgo.Session {
	session, err := mgo.Dial(os.Getenv("MONGO"))
	if err != nil {
		log.Fatal(err)
	}
	return session
}

func getRediStore() *redistore.RediStore {
	redisPass := ""
	redisURL, err := url.Parse(os.Getenv("REDIS"))
	if err != nil {
		log.Panic(err)
	}
	if redisURL.User != nil {
		redisPass, _ = redisURL.User.Password()
	}
	if redisHostPort := strings.Split(redisURL.Host,":");
		len(redisHostPort) < 2 {
		// if the host can't be split by : then append default redis port
		redisURL.Host = redisURL.Host + ":6379"
	}
	redisClients, err := strconv.Atoi(os.Getenv("REDIS_CLIENTS"))
	if err != nil {
		log.Panic(err)
		// assume undefined
		redisClients = 2
	}
	redisStore, err := redistore.NewRediStore(redisClients, "tcp",
		redisURL.Host, redisPass, []byte(os.Getenv("KEY")))
	if err != nil {
	    log.Panic(err)
	}
	return redisStore
}

func getUser(user interface{}, username string) error {
	return muser.Find(bson.M{"username": username}).One(user)
}

func getPeeves(peeves interface{}, userId bson.ObjectId) error {
	return mpeeve.Find(bson.M{"user": userId}).All(peeves)
}

func dropPeeve(peeveId string, userId bson.ObjectId) error {
	peeve := peeve{}
	err = mpeeve.Find(bson.M{"_id":bson.ObjectIdHex(peeveId), "user": userId}).One(&peeve)
	if err != nil {
		log.Panic(err)
	}
	return mpeeve.Remove(&peeve)
	// return err
}
