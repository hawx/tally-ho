package main

import (
	"flag"
	"log"

	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/handler"
)

func main() {
	var (
		port     = flag.String("port", "8080", "")
		socket   = flag.String("socket", "", "")
		me       = flag.String("me", "", "")
		dbPath   = flag.String("db", "file::memory:", "")
		baseURL  = flag.String("base-url", "http://localhost:8080/", "")
		basePath = flag.String("base-path", "/tmp/", "")
		webPath  = flag.String("web", "web", "")
	)
	flag.Parse()

	blog, err := blog.New(blog.Options{
		WebPath:  *webPath,
		BaseURL:  *baseURL,
		BasePath: *basePath,
		DbPath:   *dbPath,
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer blog.Close()

	if flag.NArg() == 1 && flag.Arg(0) == "render" {
		if err := blog.RenderAll(); err != nil {
			log.Fatal(err)
		}

		return
	}

	if *me == "" {
		log.Fatal("--me must be provided")
	}

	route.Handle("/micropub", handler.Authenticate(*me, "create", mux.Method{
		"POST": handler.Post(blog),
		"GET":  handler.Configuration(blog),
	}))

	route.Handle("/webmention", mux.Method{
		// "POST":
	})

	serve.Serve(*port, *socket, route.Default)
}
