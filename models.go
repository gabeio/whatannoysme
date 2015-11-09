package main

import (
	"log"
	"time"

	// golang.org/x/crypto
	"golang.org/x/crypto/bcrypt"

	// rethink
	r "gopkg.in/dancannon/gorethink.v1"
)

// RETHINK MODELS
type (
	// HINT: do not add spaces after commas for structure hints.
	user struct {
		Id        string    `gorethink:"id,omitempty"` //
		Username  string    `gorethink:"username"`     //
		Hash      string    `gorethink:"hash"`         //
		FirstName string    `gorethink:"firstname"`    //
		LastName  string    `gorethink:"lastname"`     //
		Email     string    `gorethink:"email"`        //
		Joined    time.Time `gorethink:"joined"`       //
	}

	peeve struct {
		Id        string    `gorethink:"id,omitempty"`      // the id of this peeve
		Root      string    `gorethink:"root"`              // original user to create the peeve
		Parent    string    `gorethink:"parent,omitempty"` // person +1 toward the origin
		// parent must be blank if this peeve is the origin
		UserId    string    `gorethink:"user"`              // the user who owns this version
		Body      string    `gorethink:"body"`              // the peeve's text
		Timestamp time.Time `gorethink:"timestamp"`         // when this version of the peeve was made
	}
)

// struct functions

func (u *user) setFirstName(firstName string) (error) {
	_, err := r.Table("users").Get(u.Id).Update(map[string]interface{}{
		"firstname": firstName,
	}).RunWrite(rethinkSession)
	return err
}

func (u *user) setLastName(lastName string) (error) {
	_, err := r.Table("users").Get(u.Id).Update(map[string]interface{}{
		"lastname": lastName,
	}).RunWrite(rethinkSession)
	return err
}

func (u *user) setPassword(password string) (error) {
	_, err := r.Table("users").Get(u.Id).Update(map[string]interface{}{
		"hash": bcryptHash(password),
	}).RunWrite(rethinkSession)
	return err
}

func (u *user) FullName() (string) {
	return u.FirstName+" "+u.LastName
}

func (p *peeve) User() (user) {
	user := user{}
	cursor, err := r.Table("users").Get(p.UserId).Run(rethinkSession)
	defer cursor.Close()
	cursor.One(user)
	if err != nil {
		log.Panic(err)
	}
	return user
}

func (p *peeve) Username() (string) {
	user := user{}
	cursor, err := r.Table("users").Get(p.UserId).Run(rethinkSession)
	defer cursor.Close()
	cursor.One(user)
	if err != nil {
		log.Panic(err)
	}
	return user.Username
}

func bcryptHash(password string) (hash string) {
	bytehash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptStrength)
	if err != nil {
		log.Panic(err) // record what "broke"
	}
	hash = string(bytehash)
	return
}
