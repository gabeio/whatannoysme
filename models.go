package main

import (
	"log"
	"time"

	// golang.org/x/crypto
	"golang.org/x/crypto/bcrypt"

	// mgo.v2 bson
	"gopkg.in/mgo.v2/bson"
)

// MONGO MODELS
type (
	user struct {
		Id          bson.ObjectId `bson:"_id"`
		Username    string        `bson:"username"`
		Hash        string        `bson:"hash"`
		FirstName   string        `bson:"firstname"`
		LastName    string        `bson:"lastname"`
		Email       string        `bson:"email"`
		Joined      time.Time     `bson:"joined"`
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

// struct functions

func (u *user) setFirstName(firstName string) (error) {
	return muser.UpdateId(u.Id, bson.M{"$set": bson.M{"firstname": firstName}})
}

func (u *user) setLastName(lastName string) (error) {
	return muser.UpdateId(u.Id, bson.M{"$set": bson.M{"lastname": lastName}})
}

func (u *user) setPassword(password string) (error) {
	return muser.UpdateId(u.Id,
		bson.M{"$set": bson.M{"hash": bcryptHash(password)}})
}

func (u *user) FullName() string {
	return u.FirstName+u.LastName
}

func (p *peeve) Username() string {
	user := user{}
	muser.FindId(p.UserId).One(&user)
	return user.Username
}

func bcryptHash(password string) (hash string) {
	bytehash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptStrength)
	if err != nil {
		log.Print(err)
	}
	hash = string(bytehash)
	return
}
