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

	blog, err := blog.New(blog.Options{
		WebPath:  *webPath,
		BaseURL:  *baseURL,
		BasePath: *basePath,
		Db:       db,
	})
	if err != nil {
		log.Println(err)
		return
	}

	if flag.NArg() == 1 && flag.Arg(0) == "render" {
		if err := blog.RenderAll(); err != nil {
			log.Println(err)
		}

		return
	}

	adminEndpoint, err := admin.Endpoint(*adminURL, *me, *secret, *webPath, blog)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/admin/", http.StripPrefix("/admin", adminEndpoint))

	micropubEndpoint, err := micropub.Endpoint(*me, blog, *mediaUploadURL)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/micropub", micropubEndpoint)

	webmentionEndpoint, _, err := webmention.Endpoint(db, blog)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/webmention", webmentionEndpoint)

	mediaEndpoint, err := media.Endpoint(*mediaPath, *mediaURL)
	if err != nil {
		log.Println("created media endpoint:", err)
		return
	}
	http.Handle("/media", mediaEndpoint)

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
