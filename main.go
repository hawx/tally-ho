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
	"hawx.me/code/tally-ho/auth"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/media"
	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/silos"
	"hawx.me/code/tally-ho/webmention"
	"hawx.me/code/tally-ho/websub"
)

func usage() {
	fmt.Println(`Usage: tally-ho [options]

  --config PATH=./config.toml
  --web DIR=web
  --db PATH=file::memory
  --media-dir DIR
  --port PORT=8080
  --socket PATH`)
}

type config struct {
	Me          string
	Name        string
	Title       string
	Description string
	BaseURL     string
	MediaURL    string

	Flickr, Twitter struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}

	Github struct {
		AccessToken string
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

	fw := &blog.FileWriter{
		MediaDir: *mediaDir,
		MediaURL: mediaURL,
	}

	var blogSilos []interface{}
	var micropubSyndicateTo []micropub.SyndicateTo

	if conf.Twitter.ConsumerKey != "" {
		twitter, err := silos.Twitter(silos.TwitterOptions{
			ConsumerKey:       conf.Twitter.ConsumerKey,
			ConsumerSecret:    conf.Twitter.ConsumerSecret,
			AccessToken:       conf.Twitter.AccessToken,
			AccessTokenSecret: conf.Twitter.AccessTokenSecret,
		}, fw)
		if err != nil {
			log.Println("WARN twitter;", err)
		} else {
			blogSilos = append(blogSilos, twitter)
			micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
				UID:  twitter.UID(),
				Name: twitter.Name(),
			})
		}
	}

	if conf.Flickr.ConsumerKey != "" {
		flickr, err := silos.Flickr(silos.FlickrOptions{
			ConsumerKey:       conf.Flickr.ConsumerKey,
			ConsumerSecret:    conf.Flickr.ConsumerSecret,
			AccessToken:       conf.Flickr.AccessToken,
			AccessTokenSecret: conf.Flickr.AccessTokenSecret,
		})
		if err != nil {
			log.Println("WARN flickr;", err)
		} else {
			blogSilos = append(blogSilos, flickr)
			micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
				UID:  flickr.UID(),
				Name: flickr.Name(),
			})
		}
	}

	if conf.Github.AccessToken != "" {
		github, err := silos.Github(silos.GithubOptions{
			AccessToken: conf.Github.AccessToken,
		})
		if err != nil {
			log.Println("WARN github;", err)
		} else {
			blogSilos = append(blogSilos, github)
			micropubSyndicateTo = append(micropubSyndicateTo, micropub.SyndicateTo{
				UID:  github.UID(),
				Name: github.Name(),
			})
		}
	}

	hubStore, err := blog.NewHubStore(db)
	if err != nil {
		log.Println("ERR blog-hub-store;", err)
		return
	}

	mediaEndpointURL, _ := url.Parse("/-/media")
	hubEndpointURL, _ := url.Parse("/-/hub")

	websubhub := websub.New(baseURL.ResolveReference(hubEndpointURL).String(), hubStore)

	b, err := blog.New(blog.Config{
		Me:          conf.Me,
		Name:        conf.Name,
		Title:       conf.Title,
		Description: conf.Description,
		BaseURL:     baseURL,
		MediaURL:    mediaURL,
		MediaDir:    *mediaDir,
		HubURL:      baseURL.ResolveReference(hubEndpointURL).String(),
	}, db, templates, websubhub, blogSilos)
	if err != nil {
		log.Println("ERR new-blog;", err)
		return
	}
	defer b.Close()

	http.Handle("/", b.Handler())

	http.Handle("/public/",
		http.StripPrefix("/public/",
			http.FileServer(
				http.Dir(filepath.Join(*webPath, "static")))))

	http.Handle("/-/micropub", micropub.Endpoint(
		b,
		conf.Me,
		baseURL.ResolveReference(mediaEndpointURL).String(),
		micropubSyndicateTo,
		fw))
	http.Handle("/-/webmention", webmention.Endpoint(b))
	http.Handle("/-/media", auth.Only(conf.Me, media.Endpoint(fw, auth.HasScope)))
	http.Handle("/-/hub", websubhub)

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
