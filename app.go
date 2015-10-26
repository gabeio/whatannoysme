package main

import (
	"os"
	"log"
	"flag"
	"time"
	"net/http"
	"math/rand"
	"html/template"

	// goji
	"github.com/zenazn/goji"

	// mgo
	"gopkg.in/mgo.v2"

	// redigo
	"github.com/garyburd/redigo/redis"
)

// cache all templates
var tem = template.Must(template.ParseGlob("templates/*.html"))

var mng *mgo.Session // mongo connection
var mdb *mgo.Database // database
var muser *mgo.Collection // user collection
var mpeeve *mgo.Collection // peeve collection

var red redis.Conn // redis connection

// random source
var ran *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// present working directory
var pwd string

var err error

func main() {
	pwd, err = os.Getwd()
	if err != nil {
		log.Panic(err)
	}
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
	goji.Get("/:username", GetPeeve)
	goji.Post("/:username", PostPeeve)
	goji.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(pwd+"/assets")))) //http.FileServer(
	flag.Set("bind", os.Getenv("SOCKET")) // set port to listen on
	goji.Serve()
}
