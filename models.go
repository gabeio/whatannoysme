package main

import (
	"time"

	// mgo.v2 bson
	"gopkg.in/mgo.v2/bson"
)

// MONGO MODELS
type (
	User struct {
		Id bson.ObjectId `bson:"_id"`
		Username string `bson:"username"`
		Hash string `bson:"hash"`
		Email string `bson:"email"`
		Joined time.Time `bson:"joined"`
	}

	Peeve struct {
		Id bson.ObjectId `bson:"_id"`
		Creator bson.ObjectId `bson:"creator"` // original user to create the peeve
		User bson.ObjectId `bson:"user"` // the user who owns this copy
		Body string `bson:"body"`
	}
)
