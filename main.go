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
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/syndicate"
	"hawx.me/code/tally-ho/webmention"
)

func usage() {
	fmt.Println(`Usage: tally-ho [options]`)
}

type config struct {
	Me          string
	Name        string
	Title       string
	Description string
	BaseURL     string

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
		log.Println("ERR decode-config;", err)
		return
	}

	db, err := blog.Open(*dbPath)
	if err != nil {
		log.Println("ERR open-blog;", err)
		return
	}
	defer db.Close()

	templates, err := blog.ParseTemplates(*webPath)
	if err != nil {
		log.Println("ERR parse-templates;", err)
		return
	}

	syndicators := map[string]syndicate.Syndicator{}

	twitter, err := syndicate.Twitter(syndicate.TwitterOptions{
		ConsumerKey:       conf.Twitter.ConsumerKey,
		ConsumerSecret:    conf.Twitter.ConsumerSecret,
		AccessToken:       conf.Twitter.AccessToken,
		AccessTokenSecret: conf.Twitter.AccessTokenSecret,
	})
	if err != nil {
		log.Println("WARN twitter;", err)
	} else {
		syndicators[syndicate.TwitterUID] = twitter
	}

	b := &blog.Blog{
		Config: blog.Config{
			Me:          conf.Me,
			Name:        conf.Name,
			Title:       conf.Title,
			Description: conf.Description,
			BaseURL:     conf.BaseURL,
		},
		DB:          db,
		Templates:   templates,
		Syndicators: syndicators,
	}

	http.Handle("/", b.Handler())

	http.Handle("/public/",
		http.StripPrefix("/public/",
			http.FileServer(
				http.Dir(filepath.Join(*webPath, "static")))))

	http.Handle("/-/micropub", micropub.Endpoint(b, conf.Me, "http://something/-/media", syndicators))
	http.Handle("/-/webmention", webmention.Endpoint(b))
	http.Handle("/-/media", http.NotFoundHandler())

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
