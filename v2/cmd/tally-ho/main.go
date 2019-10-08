package main

import (
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
	"hawx.me/code/tally-ho/v2/syndicate"
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

		twitterConsumerKey       = flag.String("twitter-consumer-key", "", "")
		twitterConsumerSecret    = flag.String("twitter-consumer-secret", "", "")
		twitterAccessToken       = flag.String("twitter-access-token", "", "")
		twitterAccessTokenSecret = flag.String("twitter-access-token-secret", "", "")

		webPath = flag.String("web", "web", "")
		dbPath  = flag.String("db", "file::memory:", "")
		port    = flag.String("port", "8080", "")
		socket  = flag.String("socket", "", "")
	)
	flag.Usage = usage
	flag.Parse()

	db, err := blog.Open(*dbPath)
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

	twitter := syndicate.Twitter(syndicate.TwitterOptions{
		ConsumerKey:       *twitterConsumerKey,
		ConsumerSecret:    *twitterConsumerSecret,
		AccessToken:       *twitterAccessToken,
		AccessTokenSecret: *twitterAccessTokenSecret,
	})

	b := &blog.Blog{
		Me:          *me,
		Name:        *name,
		Title:       *title,
		Description: *description,
		DB:          db,
		Templates:   templates,
		Twitter:     twitter,
	}

	http.Handle("/", b.Handler())

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(filepath.Join(*webPath, "static")))))
	http.Handle("/-/micropub", micropub.Endpoint(db, *me, "some-url"))
	http.Handle("/-/webmention", http.NotFoundHandler())
	http.Handle("/-/media", http.NotFoundHandler())

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
