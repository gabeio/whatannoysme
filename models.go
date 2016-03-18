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
	userModel struct {
		Id        string    `gorethink:"id,omitempty"`        //
		Username  string    `gorethink:"username,omitempty"`  //
		Hash      string    `gorethink:"hash,omitempty"`      //
		FirstName string    `gorethink:"firstname,omitempty"` //
		LastName  string    `gorethink:"lastname,omitempty"`  //
		Email     string    `gorethink:"email,omitempty"`     //
		Joined    time.Time `gorethink:"joined,omitempty"`    //
	}

	peeveModel struct {
		Id     string `gorethink:"id,omitempty"`     // the id of this peeve
		Root   string `gorethink:"root,omitempty"`   // original user to create the peeve
		Parent string `gorethink:"parent,omitempty"` // person +1 toward the origin
		// parent must be blank if this peeve is the origin
		UserId    string    `gorethink:"user,omitempty"`      // the user who owns this version
		Body      string    `gorethink:"body,omitempty"`      // the peeve's text
		Timestamp time.Time `gorethink:"timestamp,omitempty"` // when this version of the peeve was made
	}

	peeveAndUserModel struct {
		Id        string    `gorethink:"id,omitempty"`
		Username  string    `gorethink:"username,omitempty"`
		Hash      string    `gorethink:"hash,omitempty"`
		FirstName string    `gorethink:"firstname,omitempty"`
		LastName  string    `gorethink:"lastname,omitempty"`
		Email     string    `gorethink:"email,omitempty"`
		Joined    time.Time `gorethink:"joined,omitempty"`
		Root      string    `gorethink:"root,omitempty"`   // original user to create the peeve
		Parent    string    `gorethink:"parent,omitempty"` // person +1 toward the origin
		// parent must be blank if this peeve is the origin
		UserId    string    `gorethink:"user,omitempty"`      // the user who owns this version
		Body      string    `gorethink:"body,omitempty"`      // the peeve's text
		Timestamp time.Time `gorethink:"timestamp,omitempty"` // when this version of the peeve was made
	}
)

// struct functions

func (u *userModel) setFirstName(firstName string) error {
	_, err := r.Table("users").Get(u.Id).Update(userModel{
		FirstName: firstName,
	}).RunWrite(rethinkSession)
	return err
}

func (u *userModel) setLastName(lastName string) error {
	_, err := r.Table("users").Get(u.Id).Update(userModel{
		LastName: lastName,
	}).RunWrite(rethinkSession)
	return err
}

func (u *userModel) setPassword(password string) error {
	_, err := r.Table("users").Get(u.Id).Update(userModel{
		Hash: bcryptHash(password),
	}).RunWrite(rethinkSession)
	return err
}

func (u *userModel) FullName() string {
	return u.FirstName + " " + u.LastName
}

func bcryptHash(password string) (hash string) {
	bytehash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptStrength)
	if err != nil {
		log.Panic("models.bcryptHash", err) // record what "broke"
	}
	hash = string(bytehash)
	return
}
