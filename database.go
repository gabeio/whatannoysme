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
	var mgourl string
	var err error
	switch {
	case os.Getenv("MONGO") != "":
		mgourl = os.Getenv("MONGO")
	case os.Getenv("MONGO_USER") != "" &&
		os.Getenv("MONGO_PASS") != "" &&
		os.Getenv("MONGO_PORT_27017_TCP_ADDR") != "" &&
		os.Getenv("MONGO_PORT_27017_TCP_PORT") != "":
		mgourl = "mongodb://"+os.Getenv("MONGO_USER")+":"+
			os.Getenv("MONGO_PASS")+"@"+
			os.Getenv("MONGO_PORT_27017_TCP_ADDR")+":"+
			os.Getenv("MONGO_PORT_27017_TCP_PORT")
	case os.Getenv("MONGO_PORT_27017_TCP_ADDR") != "" &&
		os.Getenv("MONGO_PORT_27017_TCP_PORT") != "":
		mgourl = "mongodb://"+
			os.Getenv("MONGO_PORT_27017_TCP_ADDR")+":"+
			os.Getenv("MONGO_PORT_27017_TCP_PORT")
	}
	session, err := mgo.Dial(mgourl)
	if err != nil {
		log.Fatal(err)
	}
	return session
}

func getRediStore() *redistore.RediStore {
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
		if redisHostPortArray := strings.Split(redisURL.Host,":");
			len(redisHostPortArray) < 2 {
			// if the host can't be split by : then append default redis port
			redisHostPort += ":6379"
		}
	case os.Getenv("REDIS_PASS") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_PORT") != "":
		redisHostPort = os.Getenv("REDIS_PORT_6379_TCP_ADDR")+":"+
			os.Getenv("REDIS_PORT_6379_TCP_PORT")
		redisPassword = os.Getenv("REDIS_PASS")
	case os.Getenv("REDIS_PORT_6379_TCP_ADDR") != "" &&
		os.Getenv("REDIS_PORT_6379_TCP_PORT") != "":
		redisHostPort = os.Getenv("REDIS_PORT_6379_TCP_ADDR")+":"+
			os.Getenv("REDIS_PORT_6379_TCP_PORT")
	}
	redisStore, err := redistore.NewRediStore(redisClients, "tcp",
		redisHostPort, redisPassword, []byte(os.Getenv("KEY")))
	if err != nil {
	    log.Panic(err)
	}
	return redisStore
}

func getUser(username string, user interface{}) error {
	return muser.Find(bson.M{"username": username}).One(user)
}

func searchUser(query string, users interface{}) error {
	muser.EnsureIndexKey("username")
	err = muser.Find(bson.M{"$text": bson.M{"$search": query}}).All(users)
	if err != nil {
		return err
	}
	// only return nil if no errors
	return nil
}

func getPeeves(userId bson.ObjectId, peeves interface{}) error {
	return mpeeve.Find(bson.M{"user": userId}).All(peeves)
}

func searchPeeve(query string, peeves interface{}) error {
	muser.EnsureIndexKey("body")
	err = mpeeve.Find(bson.M{"$text": bson.M{"$search": query}}).All(peeves)
	if err != nil {
		return err
	}
	// only return nil if no errors
	return nil
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
