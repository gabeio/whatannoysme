package main

import (
	"io"
	"os"
	"log"
	"flag"
	"time"
	"strconv"
	"net/http"
	"math/rand"
	"html/template"

	// goji
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	// mgo
	"gopkg.in/mgo.v2"
	// redigo
	"github.com/garyburd/redigo/redis"
)

var tem = template.Must(template.ParseGlob("templates/*.html")) // cache all templates

var mng *mgo.Session // mongo connection

var red redis.Conn // redis connection

var ran *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	var err error
	mng, err = mgo.Dial(os.Getenv("MONGO"))
	if err != nil {
		log.Fatal(err)
	}
	red, err = redis.DialURL(os.Getenv("REDIS"))
	if err != nil {
		log.Fatal(err)
	}
	goji.Use(HtmlText) // Only Serve Html Back
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
	defer mng.Close()
	defer red.Close()
}

func Root(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "index", nil) // only serve index.html
}

func Login(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "login", nil)
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1048576) // translate multipart 1MB limit

	log.Print(r.Form["username"][0])
	log.Print(r.Form["password"][0])
	io.WriteString(w,"ok")
}

func Signup(w http.ResponseWriter, r *http.Request) {
	tem.ExecuteTemplate(w, "signup", nil)
}

func PostSignup(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1048576) // translate multipart 1MB limit
	if r.Form["username"] == nil {
		tem.ExecuteTemplate(w, "signup", nil)
	}else if r.Form["password"] == nil {
		randIssue := strconv.FormatInt(rand.Int63(), 10)
		io.WriteString(w, randIssue)
		log.Print(randIssue)
	}else if r.Form["password"][0] != r.Form["password"][1] {
		randIssue := strconv.FormatInt(rand.Int63(), 10)
		io.WriteString(w, randIssue)
		log.Print(randIssue)
	}else{
		//signup
		io.WriteString(w, "OKAY!")
	}
}

func Random(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, strconv.FormatInt(rand.Int63(), 10))
}

func GetUser(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Print(c.URLParams["username"])
	tem.ExecuteTemplate(w, "user", c.URLParams["username"])
}

func PostUser(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1048576) // translate multipart
	if r.Form["body"] == nil {
		// log.Print(w.Body)
		// log.Print(r)
		// w.Body.Close()
		io.WriteString(w,"") // closest thing to .close()
	}
	io.WriteString(w,r.Form["body"][0])
}
