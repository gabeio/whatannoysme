package main

import (
	"io"
	"log"
	"flag"
	"net/http"
	"html/template"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

var t = template.Must(template.ParseGlob("templates/*.html")) // cache all templates

func main() {
	goji.Use(HtmlText) // Only Serve Html Back
	goji.Get("/", Root)
	goji.Get("/login", Login)
	goji.Post("/login", PostLogin)
	goji.Get("/:username", User)
	goji.Post("/:username", PostUser)
	flag.Set("bind", ":9000") // set port to listen on
	goji.Serve()
}

func Root(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "index", nil) // only serve index.html
}

func Login(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "login", nil)
}

func PostLogin(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1024) // translate multipart
	log.Print(len(r.Form))
	// for i,e := range r.Form {
	// 	log.Print(i)
	// 	log.Print(e)
	// }
	log.Print(r.Form["username"][0])
	log.Print(r.Form["password"][0])
	io.WriteString(w,"ok")
}

func User(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Print(c.URLParams["username"])
	io.WriteString(w,c.URLParams["username"])
}

func PostUser(c web.C, w http.ResponseWriter, r *http.Request) {
	r.ParseForm() // translate form
	r.ParseMultipartForm(1024) // translate multipart

	log.Print(c.URLParams["username"])
	io.WriteString(w,c.URLParams["username"])
}
