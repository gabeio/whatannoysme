package main

import (
	"os"
	"log"
	"flag"
	"html/template"

	// goji
	"github.com/zenazn/goji"

	// rethink
	r "github.com/dancannon/gorethink"

	// redis
	redis "gopkg.in/boj/redistore.v1"
)

var (
	// load & cache all templates
	temps *template.Template = template.Must(template.ParseGlob(os.Getenv("GOPATH") +
		"/src/github.com/gabeio/whatannoysme/templates/*.html"))
	// session name
	sessionName = "wam"
	// rethink session
	rethinkSession *r.Session
	// redis store
	redisStore *redis.RediStore
	// global bcrypt strength
	bcryptStrength int = 12
	// present working directory
	pwd string
	// errors
	err error
	errs = make(chan error)
)

func main() {
	// check all necessary Environmental variables are present
	checkEnvs()

	// get and save the present working directory
	pwd, err = os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	// start rethink db connect
	rethinkChannel := make(chan *r.Session)
	go getRethinkSession(rethinkChannel)

	// start redis session connect
	redisChannel := make(chan *redis.RediStore)
	go getRediStore(redisChannel)

	// save database connections to "global" variables
	rethinkSession = <-rethinkChannel
	redisStore = <-redisChannel

	// close channels
	close(rethinkChannel)
	close(redisChannel)

	// defer close database connections
	defer rethinkSession.Close()
	defer redisStore.Close()

	// max session length
	redisStore.SetMaxAge(7 * 24 * 3600) // 7 days

	// goji
	goji.Use(textHtml) // serve text/html
	goji.Post("/too", MeTooPeeve)
	goji.Get("/", IndexTemplate)
	goji.Get("/login", LoginTemplate)
	goji.Post("/login", Login)
	goji.Get("/logout", Logout) // REMOVE THIS!
	goji.Post("/logout", Logout)
	goji.Get("/signup", SignupTemplate)
	goji.Post("/signup", CreateUser)
	goji.Get("/search", Search)
	goji.Post("/search", Search)
	goji.Get("/:username", GetPeeves)
	goji.Post("/:username/create", CreatePeeve)
	goji.Post("/:username/delete", DeletePeeve)
	goji.Get("/:username/settings", SettingsTemplate)
	goji.Post("/:username/settings", Settings)
	flag.Set("bind", getSocket()) // set port to listen on
	goji.Serve()
}

func getSocket() string {
	if os.Getenv("SOCKET") != "" {
		return os.Getenv("SOCKET")
	} else {
		return ":8080"
	}
}

func checkEnvs() {
	switch {
	case os.Getenv("SOCKET") == "":
		log.Print("SOCKET variable undefined assuming 8080")
	case os.Getenv("RETHINK") == "" &&
		os.Getenv("RETHINK_PORT_28015_TCP_ADDR") == "" &&
		os.Getenv("RETHINK_PORT_28015_TCP_PORT") == "":
		log.Fatal("Environmental RETHINK variable required")
	case os.Getenv("RETHINK_DB") == "":
		log.Print("RETHINK_DB variable undefined")
	case os.Getenv("REDIS") == "":
		log.Fatal("Environmental REDIS variable required")
	case os.Getenv("REDIS_CLIENTS") == "":
		log.Fatal("Environmental REDIS_CLIENTS variable required")
	case os.Getenv("KEY") == "":
		log.Fatal("Environmental KEY variable required")
	}
}
