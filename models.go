package main

import (
	"time"

	// mgo.v2 bson
	"gopkg.in/mgo.v2/bson"
)

// MONGO MODELS
type (
	user struct {
		Id        bson.ObjectId `bson:"_id"`
		Username  string        `bson:"username"`
		Hash      string        `bson:"hash"`
		FirstName string        `bson:"firstname"`
		LastName  string        `bson:"lastname"`
		Email     string        `bson:"email"`
		Joined    time.Time     `bson:"joined"`
	}

	peeve struct {
		Id        bson.ObjectId `bson:"_id"`
		Root      bson.ObjectId `bson:"root"`             // original user to create the peeve
		Parent    bson.ObjectId `bson:"parent,omitempty"` // person +1 toward the root
		UserId    bson.ObjectId `bson:"user"`             // the user who owns this copy/peeve
		Body      string        `bson:"body"`
		Timestamp time.Time     `bson:"timestamp"`
	}
)

func (u *user) FullName() string {
	return u.FirstName+u.LastName
}

func (p *peeve) Username() string {
	user := user{}
	muser.FindId(p.UserId).One(&user)
	return user.Username
}
