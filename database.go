package main

import (
	"os"
	"log"

	// mgo
	"gopkg.in/mgo.v2"

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
