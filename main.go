package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/admin"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/media"
	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/webmention"
	"hawx.me/code/tally-ho/writer"
)

func main() {
	var (
		port           = flag.String("port", "8080", "")
		socket         = flag.String("socket", "", "")
		me             = flag.String("me", "", "")
		dbPath         = flag.String("db", "file::memory:", "")
		baseURL        = flag.String("base-url", "http://localhost:8080/", "")
		basePath       = flag.String("base-path", "/tmp/", "")
		mediaURL       = flag.String("media-url", "http://localhost:8080/_media/", "")
		mediaPath      = flag.String("media-path", "/tmp/", "")
		adminURL       = flag.String("admin-url", "http://localhost:8080/admin/", "")
		mediaUploadURL = flag.String("media-upload-url", "http://localhost:8080/media", "")
		webPath        = flag.String("web", "web", "")
		secret         = flag.String("secret", "", "")
	)
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	fw, err := writer.NewFileWriter(*basePath, *baseURL)
	if err != nil {
		log.Println(err)
		return
	}

	templates, err := blog.ParseTemplates(*webPath)
	if err != nil {
		log.Println(err)
		return
	}

	blog := &blog.Blog{
		BaseURL:    *baseURL,
		FileWriter: fw,
		Templates:  templates,
	}

	micropubEndpoint, mr, err := micropub.Endpoint(db, *me, blog, *mediaUploadURL, fw)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/micropub", micropubEndpoint)

	blog.Entries = mr

	adminEndpoint, err := admin.Endpoint(*adminURL, *me, *secret, *webPath, mr, templates)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/admin/", http.StripPrefix("/admin", adminEndpoint))

	webmentionEndpoint, wr, err := webmention.Endpoint(db, mr, blog)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/webmention", webmentionEndpoint)

	blog.Mentions = wr

	mediaEndpoint, err := media.Endpoint(*mediaPath, *mediaURL)
	if err != nil {
		log.Println("created media endpoint:", err)
		return
	}
	http.Handle("/media", mediaEndpoint)

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
