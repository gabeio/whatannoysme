package main

import (
	"time"

	// mgo.v2 bson
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id bson.ObjectId `bson:"_id"`
	Username string `bson:"username"`
	Hash string `bson:"hash"`
	Email string `bson:"email"`
	Joined time.Time `bson:"joined"`
}

type Peeve struct {
	Id bson.ObjectId `bson:"_id"`
	UserId bson.ObjectId `bson:"uid"`
	Body string `bson:"body"`
}

func (u User) SetUsername(p Peeve) (int) {
	return 0
}
