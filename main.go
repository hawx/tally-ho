package main

import (
	"flag"
	"log"

	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/config"
	"hawx.me/code/tally-ho/data"
	"hawx.me/code/tally-ho/handler"
	"hawx.me/code/tally-ho/renderer"
)

func main() {
	var (
		port     = flag.String("port", "8080", "")
		socket   = flag.String("socket", "", "")
		me       = flag.String("me", "", "")
		dbPath   = flag.String("db", "file::memory:", "")
		baseURL  = flag.String("base-url", "http://localhost:8080/", "")
		basePath = flag.String("base-path", "/tmp/", "")
	)
	flag.Parse()

	config, err := config.New(*baseURL, *basePath)
	if err != nil {
		log.Fatal(err)
	}

	render, err := renderer.New(config, "web/template/*.gotmpl")
	if err != nil {
		log.Fatal(err)
	}

	store, err := data.Open(*dbPath, config)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	if flag.NArg() == 1 && flag.Arg(0) == "render" {
		if err := render.All(store); err != nil {
			log.Fatal(err)
		}

		return
	}

	if *me == "" {
		log.Fatal("--me must be provided")
	}

	route.Handle("/micropub", handler.Authenticate(*me, "create", mux.Method{
		"POST": handler.Post(store, render, config),
		"GET":  handler.Configuration(store, config),
	}))

	route.Handle("/webmention", mux.Method{
		// "POST":
	})

	serve.Serve(*port, *socket, route.Default)
}
