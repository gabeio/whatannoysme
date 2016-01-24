package main

import (
	"html/template"
	"log"
	"os"

	// echo
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"

	// secure
	secure "gopkg.in/unrolled/secure.v1"

	// rethink
	rethink "gopkg.in/dancannon/gorethink.v1"

	// redis
	redis "gopkg.in/boj/redistore.v1"
)

var (
	// load & cache all templates
	temps = &Template{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	// session name
	sessionName = "wam"
	// rethink session
	rethinkSession *rethink.Session
	// redis store
	redisStore *redis.RediStore
	// global bcrypt strength
	bcryptStrength int = 12
	// present working directory
	pwd string
	// errors
	err  error
	errs = make(chan error)
	// security settings
	securemw = secure.New(secure.Options{
		AllowedHosts:       []string{"whatannoys.me", "www.whatannoys.me"},
		SSLProxyHeaders:    map[string]string{"X-Forwarded-Proto": "https"},
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		ContentSecurityPolicy: "default-src 'self';" +
			"script-src 'self' 'unsafe-inline' cdnjs.cloudflare.com maxcdn.bootstrapcdn.com;" +
			"style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com maxcdn.bootstrapcdn.com;" +
			"img-src 'none';"+
			"connect-src 'none';"+
			"font-src maxcdn.bootstrapcdn.com;"+
			"object-src 'none';"+
			"media-src 'none';"+
			"frame-src 'none';"+
			"",
	})
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
	rethinkChannel := make(chan *rethink.Session)
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

	// echo setup
	e := echo.New()
	// e.Debug()
	e.Use(securemw.Handler)
	e.Use(mw.Recover())
	e.HTTP2(true)
	e.SetRenderer(temps) // connect render(er)
	e.Post("/too", MeTooPeeve)
	e.Get("/", IndexTemplate)
	e.Get("/login", LoginTemplate)
	e.Post("/login", Login)
	e.Get("/logout", Logout) // REMOVE THIS!
	e.Post("/logout", Logout)
	e.Get("/signup", SignupTemplate)
	e.Post("/signup", CreateUser)
	e.Get("/search", Search)
	e.Post("/search", Search)
	e.Get("/:username", GetPeeves)
	e.Post("/:username/create", CreatePeeve)
	e.Post("/:username/delete", DeletePeeve)
	e.Get("/:username/settings", SettingsTemplate)
	e.Post("/:username/settings", Settings)
	e.Run(getSocket())
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
