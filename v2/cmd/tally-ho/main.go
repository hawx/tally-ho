package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"github.com/BurntSushi/toml"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/v2/blog"
	"hawx.me/code/tally-ho/v2/micropub"
	"hawx.me/code/tally-ho/v2/syndicate"
	"hawx.me/code/tally-ho/v2/webmention"
)

func usage() {
	fmt.Println(`Usage: tally-ho [options]`)
}

type config struct {
	Me          string
	Name        string
	Title       string
	Description string

	Twitter struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}
}

func main() {
	var (
		configPath = flag.String("config", "./config.toml", "")
		webPath    = flag.String("web", "web", "")
		dbPath     = flag.String("db", "file::memory:", "")
		port       = flag.String("port", "8080", "")
		socket     = flag.String("socket", "", "")
	)
	flag.Usage = usage
	flag.Parse()

	var conf config
	if _, err := toml.DecodeFile(*configPath, &conf); err != nil {
		log.Println(err)
		return
	}

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
		ConsumerKey:       conf.Twitter.ConsumerKey,
		ConsumerSecret:    conf.Twitter.ConsumerSecret,
		AccessToken:       conf.Twitter.AccessToken,
		AccessTokenSecret: conf.Twitter.AccessTokenSecret,
	})

	b := &blog.Blog{
		Me:          conf.Me,
		Name:        conf.Name,
		Title:       conf.Title,
		Description: conf.Description,
		DB:          db,
		Templates:   templates,
		Twitter:     twitter,
	}

	http.Handle("/", b.Handler())

	http.Handle("/public/",
		http.StripPrefix("/public/",
			http.FileServer(
				http.Dir(filepath.Join(*webPath, "static")))))

	http.Handle("/-/micropub", micropub.Endpoint(b, conf.Me, "http://something/-/media"))
	http.Handle("/-/webmention", webmention.Endpoint(b))
	http.Handle("/-/media", http.NotFoundHandler())

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
