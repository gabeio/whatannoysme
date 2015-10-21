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
	goji.Use(HtmlText)
	goji.Get("/", Root)
	// goji.Get("/login", Login)
	// goji.Post("/login", PostLogin)
	goji.Get("/:username", User)
	flag.Set("bind", ":9000") // set port to listen on
	goji.Serve()
}

func Root(w http.ResponseWriter, r *http.Request) {
	t.ExecuteTemplate(w, "index.html", nil) // only serve index.html
}

func User(c web.C, w http.ResponseWriter, r *http.Request) {
	log.Print(c.URLParams["username"])
	io.WriteString(w,c.URLParams["username"])
}
