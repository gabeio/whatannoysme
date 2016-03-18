package main

import (
	"log"
	"os"

	// gin
	"github.com/gin-gonic/gin"

	// gzip
	gzip "github.com/gin-gonic/contrib/gzip"

	// secure
	secure "github.com/gin-gonic/contrib/secure"

	// rethink
	rethink "gopkg.in/dancannon/gorethink.v1"

	// redis
	redis "gopkg.in/boj/redistore.v1"

	// endless
	"github.com/fvbock/endless"
)

var (
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
	// err  error
	// errs := make(chan error)
)

func main() {
	// check all necessary Environmental variables are present
	checkEnvs()

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

	// router setup
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(secure.Secure(secure.Options{
		AllowedHosts:       []string{"localhost:8090", "whatannoys.me", "www.whatannoys.me", "direct.whatannoys.me"},
		SSLProxyHeaders:    map[string]string{"X-Forwarded-Proto": "https"},
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		ContentSecurityPolicy: "default-src 'self';" +
			"script-src 'self' 'unsafe-inline' cdnjs.cloudflare.com maxcdn.bootstrapcdn.com;" +
			"style-src 'self' 'unsafe-inline' cdnjs.cloudflare.com maxcdn.bootstrapcdn.com;" +
			"img-src 'none';" +
			"connect-src 'none';" +
			"font-src maxcdn.bootstrapcdn.com;" +
			"object-src 'none';" +
			"media-src 'none';" +
			"frame-src 'none';" +
			"",
	}))
	r.LoadHTMLGlob("templates/*")
	r.GET("/", IndexTemplate)
	r.POST("/too", MeTooPeeve)
	r.GET("/login", LoginTemplate)
	r.POST("/login", Login)
	r.GET("/logout", Logout) // REMOVE THIS!
	r.POST("/logout", Logout)
	r.GET("/signup", SignupTemplate)
	r.POST("/signup", CreateUser)
	r.GET("/search", Search)
	r.POST("/search", Search)
	r.GET("/settings", SettingsTemplate)
	r.POST("/settings", Settings)
	u := r.Group("/u/:username")
	{
		u.POST("/create", CreatePeeve)
		u.POST("/delete", DeletePeeve)
		u.GET("", GetPeeves)
	}
	endless.ListenAndServe(getSocket(), r)
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
