package main

import (
	"os"
	"log"

	// mgo
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// redigo
	"github.com/garyburd/redigo/redis"
)

func getMgoSession() *mgo.Session {
	session, err := mgo.Dial(os.Getenv("MONGO"))
	if err != nil {
		log.Fatal(err)
	}
	return session
}

func getRedisConn() redis.Conn {
	connection, err := redis.DialURL(os.Getenv("REDIS"))
	if err != nil {
		log.Fatal(err)
	}
	return connection
}

func getUser(user interface{}, username string) error {
	err = muser.Find(bson.M{"username": username}).One(user)
	return err
}

func getPeeves(peeves interface{}, userId bson.ObjectId) error {
	err = mpeeve.Find(bson.M{"user": userId}).All(peeves)
	return err
}

func dropPeeve(peeveId bson.ObjectId, userId bson.ObjectId) error {
	peeve := peeve{}
	mpeeve.Find(bson.M{"_id":peeveId, "user": userId}).One(&peeve)
	err = mpeeve.Remove(&peeve)
	return err
}
