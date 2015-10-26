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
	bcryptStrength int = 12
	pwd string // present working directory
	err error
)

func main() {
	pwd, err = os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	// db connects
	msess = getMgoSession()
	defer msess.Close()
	rconn = getRedisConn()
	defer rconn.Close()
	// further db insides
	mdb = msess.DB(os.Getenv("MONGO_DB"))
	muser = mdb.C("user")
	mpeeve = mdb.C("peeve")
	goji.Use(textHtml) // serve text/html
	goji.Get("/", IndexTemplate)
	goji.Get("/login", LoginTemplate)
	goji.Post("/login", Login)
	goji.Get("/signup", SignupTemplate)
	goji.Post("/signup", CreateUser)
	goji.Get("/:username", GetPeeves)
	goji.Post("/:username", CreatePeeve)
	flag.Set("bind", os.Getenv("SOCKET")) // set port to listen on
	goji.Serve()
}
