package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"github.com/BurntSushi/toml"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/auth"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/internal/page"
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
	Me       string
	BaseURL  string
	MediaURL string

	Context page.Context

	Flickr struct {
		ConsumerKey       string
		ConsumerSecret    string
		AccessToken       string
		AccessTokenSecret string
	}

	Github struct {
		AccessToken string
	}
}

type configLink struct {
	Name string
	URL  string
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

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	var conf config
	if _, err := toml.DecodeFile(*configPath, &conf); err != nil {
		logger.Error("config could not be decoded", slog.Any("err", err))
		return
	}

	baseURL, err := url.Parse(conf.BaseURL)
	if err != nil {
		logger.Error("base url invalid", slog.Any("err", err))
		return
	}

	mediaURL, err := url.Parse(conf.MediaURL)
	if err != nil {
		logger.Error("media url invalid", slog.Any("err", err))
		return
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logger.Error("error opening sqlite file", slog.String("path", *dbPath), slog.Any("err", err))
		return
	}

	fw := &blog.FileWriter{
		MediaDir: *mediaDir,
		MediaURL: mediaURL,
	}

	var blogSilos []any
	var micropubSyndicateTo []micropub.SyndicateTo

	if conf.Flickr.ConsumerKey != "" {
		flickr, err := silos.Flickr(silos.FlickrOptions{
			ConsumerKey:       conf.Flickr.ConsumerKey,
			ConsumerSecret:    conf.Flickr.ConsumerSecret,
			AccessToken:       conf.Flickr.AccessToken,
			AccessTokenSecret: conf.Flickr.AccessTokenSecret,
		})
		if err != nil {
			logger.Warn("flickr", slog.Any("err", err))
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
			logger.Warn("github", slog.Any("err", err))
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
		logger.Error("problem initialising hub store", slog.Any("err", err))
		return
	}

	mediaEndpointURL, _ := url.Parse("/-/media")
	hubEndpointURL, _ := url.Parse("/-/hub")

	websubhub := websub.New(baseURL.ResolveReference(hubEndpointURL).String(), hubStore)

	b, err := blog.New(logger, blog.Config{
		Me:       conf.Me,
		BaseURL:  baseURL,
		MediaURL: mediaURL,
		HubURL:   baseURL.ResolveReference(hubEndpointURL).String(),
	}, conf.Context, db, websubhub, blogSilos)
	if err != nil {
		logger.Error("problem initialising blog", slog.Any("err", err))
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
