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
