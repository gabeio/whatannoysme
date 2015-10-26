package main

import (
	"os"
	"log"
	"flag"
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
var (
	// cache all templates
	temps *template.Template =
		template.Must(template.ParseGlob("templates/*.html"))
	msess *mgo.Session // mongo connection
	mdb *mgo.Database // database
	muser *mgo.Collection // user collection
	mpeeve *mgo.Collection // peeve collection
	rconn redis.Conn // redis connection
	// random source
	ran *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	pwd string // present working directory
	err error
)

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
	goji.Use(TextHtml) // serve text/html
	goji.Get("/", Root)
	goji.Get("/login", Login)
	goji.Post("/login", PostLogin)
	goji.Get("/random", Random)
	goji.Get("/signup", Signup)
	goji.Post("/signup", PostSignup)
	goji.Get("/:username", GetPeeve)
	goji.Post("/:username", PostPeeve)
	flag.Set("bind", os.Getenv("SOCKET")) // set port to listen on
	goji.Serve()
}
