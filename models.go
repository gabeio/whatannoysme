package main

import (
	"time"

	// mgo.v2 bson
	"gopkg.in/mgo.v2/bson"
)

// MONGO MODELS
type (
	user struct {
		Id bson.ObjectId `bson:"_id"`
		Username string `bson:"username"`
		Hash string `bson:"hash"`
		Email string `bson:"email"`
		Joined time.Time `bson:"joined"`
	}

	peeve struct {
		Id bson.ObjectId `bson:"_id"`
		Root bson.ObjectId `bson:"root"` // original user to create the peeve
		Parent bson.ObjectId `bson:"parent,omitempty"` // person +1 toward the root
		User bson.ObjectId `bson:"user"` // the user who owns this copy/peeve
		Username bson.ObjectId
		Body string `bson:"body"`
		Timestamp time.Time `bson:"timestamp"`
	}
)
