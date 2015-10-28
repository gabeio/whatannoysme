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

	// redis
	// "github.com/garyburd/redigo/redis"
	"gopkg.in/boj/redistore.v1"
)

var (
	// cache all templates
	temps *template.Template =
		template.Must(template.ParseGlob("templates/*.html"))
	msess *mgo.Session // mongo connection
	mdb *mgo.Database // database
	muser *mgo.Collection // user collection
	mpeeve *mgo.Collection // peeve collection
	rstore *redistore.RediStore // redis store
	// random source
	ran *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	bcryptStrength int = 12
	pwd string // present working directory
	err error
)

func main() {
	checkEnvs()
	pwd, err = os.Getwd()
	if err != nil {
		log.Panic(err)
	}
	// mongo db connect
	msess = getMgoSession()
	defer msess.Close()
	// redis session connect
	rstore = getRediStore()
	defer rstore.Close()
	rstore.SetMaxAge(7*24*3600) // 7 days
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
	goji.Post("/search", Search)
	goji.Get("/:username", GetPeeves)
	goji.Post("/:username/create", CreatePeeve)
	goji.Post("/:username/delete", DeletePeeve)
	flag.Set("bind", os.Getenv("SOCKET")) // set port to listen on
	goji.Serve()
}

func checkEnvs() {
	switch {
	case os.Getenv("SOCKET") == "":
		log.Fatal("Environmental SOCKET variable required")
	case os.Getenv("REDIS") == "":
		log.Fatal("Environmental REDIS variable required")
	case os.Getenv("REDIS_CLIENTS") == "":
		log.Fatal("Environmental REDIS_CLIENTS variable required")
	case os.Getenv("KEY") == "":
		log.Fatal("Environmental KEY variable required")
	case os.Getenv("MONGO") == "":
		log.Fatal("Environmental MONGO variable required")
	case os.Getenv("MONGO_DB") == "":
		log.Print("Environmental MONGO_DB variable suggested")
	}
}
