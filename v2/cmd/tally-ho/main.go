package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/v2/blog"
	"hawx.me/code/tally-ho/v2/micropub"
)

func usage() {
	fmt.Println(`Usage: tally-ho [options]`)
}

func main() {
	var (
		me          = flag.String("me", "", "")
		name        = flag.String("name", "", "")
		title       = flag.String("title", "", "")
		description = flag.String("description", "", "")

		webPath = flag.String("web", "web", "")
		dbPath  = flag.String("db", "file::memory:", "")
		port    = flag.String("port", "8080", "")
		socket  = flag.String("socket", "", "")
	)
	flag.Usage = usage
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	templates, err := blog.ParseTemplates(*webPath)
	if err != nil {
		log.Println(err)
		return
	}

	micropubEndpoint, micropubReader, err := micropub.Endpoint(db, *me, "some-url")
	if err != nil {
		log.Println(err)
		return
	}

	b := &blog.Blog{
		Me:          *me,
		Name:        *name,
		Title:       *title,
		Description: *description,
		Db:          db,
		Templates:   templates,
		Posts:       micropubReader,
	}

	http.Handle("/", b.Handler())

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(filepath.Join(*webPath, "static")))))
	http.Handle("/-/micropub", micropubEndpoint)
	http.Handle("/-/webmention", http.NotFoundHandler())
	http.Handle("/-/media", http.NotFoundHandler())

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
