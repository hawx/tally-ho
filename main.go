package main

import (
	"flag"
	"log"

	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
)

func main() {
	var (
		port   = flag.String("port", "8080", "")
		socket = flag.String("socket", "", "")
		me     = flag.String("me", "", "")
	)
	flag.Parse()

	if *me == "" {
		log.Fatal("--me must be provided")
	}

	route.Handle("/micropub", mux.Method{
	// "POST": handler.Authenticate(*me, "create", handler.Post()),
	})

	route.Handle("/webmention", mux.Method{
	// "POST":
	})

	serve.Serve(*port, *socket, route.Default)
}
