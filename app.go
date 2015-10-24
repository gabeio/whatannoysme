package main

import (
	"os"
	"flag"
	"math"
	"time"
	"math/rand"
	"html/template"

	// goji
	"github.com/zenazn/goji"

	// mgo
	"gopkg.in/mgo.v2"

	// redigo
	"github.com/garyburd/redigo/redis"
)

var tem = template.Must(template.ParseGlob("templates/*.html")) // cache all templates

var mng *mgo.Session // mongo connection
var mdb *mgo.Database // database
var muser *mgo.Collection // user collection
var mpeeve *mgo.Collection // peeve collection

var red redis.Conn // redis connection

var ran *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano())) // random source

var err error

func main() {
	// db connects
	mng = getMgoSession()
	defer mng.Close() // try to always close
	red = getRedisConn()
	defer red.Close() // try to always close
	// further db insides
	mdb = mng.DB(os.Getenv("MONGO_DB"))
	muser = mdb.C("user")
	mpeeve = mdb.C("peeve")
	goji.Use(HtmlText) // only serve html/text
	goji.Get("/", Root)
	goji.Get("/login", Login)
	goji.Post("/login", PostLogin)
	goji.Get("/random", Random)
	goji.Get("/signup", Signup)
	goji.Post("/signup", PostSignup)
	goji.Get("/:username", GetUser)
	goji.Post("/:username", PostUser)
	flag.Set("bind", os.Getenv("SOCKET")) // set port to listen on
	goji.Serve()
}
