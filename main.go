package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"github.com/BurntSushi/toml"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/media"
	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/silos"
	"hawx.me/code/tally-ho/webmention"
	"hawx.me/code/tally-ho/websub"
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
	MediaURL    string

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
		mediaDir   = flag.String("media-dir", "", "")
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

	templates, err := blog.ParseTemplates(*webPath)
	if err != nil {
		log.Println("ERR parse-templates;", err)
		return
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Printf("ERR sql-open path=%s; %v\n", *dbPath, err)
		return
	}

	blogCiters := []blog.Citer{}
	blogSyndicators := map[string]blog.Syndicator{}
	micropubSyndicators := map[string]micropub.Syndicator{}

	twitter, err := silos.Twitter(silos.TwitterOptions{
		ConsumerKey:       conf.Twitter.ConsumerKey,
		ConsumerSecret:    conf.Twitter.ConsumerSecret,
		AccessToken:       conf.Twitter.AccessToken,
		AccessTokenSecret: conf.Twitter.AccessTokenSecret,
	})
	if err != nil {
		log.Println("WARN twitter;", err)
	} else {
		blogCiters = append(blogCiters, twitter)
		blogSyndicators[silos.TwitterUID] = twitter
		micropubSyndicators[silos.TwitterUID] = twitter
	}

	baseURL, err := url.Parse(conf.BaseURL)
	if err != nil {
		log.Println("ERR base-url-invalid;", err)
		return
	}

	mediaURL, err := url.Parse(conf.MediaURL)
	if err != nil {
		log.Println("ERR media-url-invalid;", err)
		return
	}

	hubStore, err := blog.NewHubStore(db)
	if err != nil {
		log.Println("ERR blog-hub-store;", err)
		return
	}

	websubhub := websub.New(baseURL.String(), hubStore)

	b, err := blog.New(blog.Config{
		Me:          conf.Me,
		Name:        conf.Name,
		Title:       conf.Title,
		Description: conf.Description,
		BaseURL:     baseURL,
		MediaURL:    mediaURL,
		MediaDir:    *mediaDir,
	}, db, templates, blogSyndicators, blogCiters, websubhub)
	if err != nil {
		log.Println("ERR new-blog;", err)
		return
	}
	defer b.Close()

	mediaEndpointURL, _ := url.Parse("/-/media")

	http.Handle("/", b.Handler())

	http.Handle("/public/",
		http.StripPrefix("/public/",
			http.FileServer(
				http.Dir(filepath.Join(*webPath, "static")))))

	http.Handle("/-/micropub", micropub.Endpoint(
		b,
		conf.Me,
		baseURL.ResolveReference(mediaEndpointURL).String(),
		micropubSyndicators,
		b))
	http.Handle("/-/webmention", webmention.Endpoint(b))
	http.Handle("/-/media", media.Endpoint(conf.Me, b))
	http.Handle("/-/hub", websubhub)

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
