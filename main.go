package main

import (
	"database/sql"
	"flag"
	"fmt"
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

func usage() {
	fmt.Println(`Usage: tally-ho [options]

  Tally-ho is a blog.

 META
   --me URL
      Full URL to your website, only this user will have access to
      create new posts. Also used as the author's URL in templates.

   --name NAME
   --title TITLE
   --description DESC
      Set values for the author's name and the blog.

 PATH REWRITING
   You need to tell tally-ho how it will be served and where to write
   things. All URLs and PATHs must end with a '/'.

   --blog-url URL
   --blog-path PATH
      Set the URL that the blog will be served at and the directory to
      write posts to.

   --media-url URL
   --media-path PATH
      Set the URL that media files will be served at and the directory
      to write them to.

   --admin-url URL
      Set the URL that the admin interface will be served at.

   --media-upload-url URL
      Set the URL that this application will be served at.

 DATA
   --web PATH
      The path to the 'web' directory of this project.

   --db PATH
      The path to the database, otherwise an in-memory store is used.

   --secret STRING
      Secret to secure cookies with, base64 encoded.

 SERVE
   --port PORT='8080'
      Serve on given port.

   --socket SOCK
      Serve at given socket, instead.`)
}

func main() {
	var (
		me          = flag.String("me", "", "")
		name        = flag.String("name", "", "")
		title       = flag.String("title", "", "")
		description = flag.String("description", "", "")

		blogURL        = flag.String("blog-url", "http://localhost:8080/", "")
		blogPath       = flag.String("blog-path", "/tmp/", "")
		mediaURL       = flag.String("media-url", "http://localhost:8080/_media/", "")
		mediaPath      = flag.String("media-path", "/tmp/", "")
		adminURL       = flag.String("admin-url", "http://localhost:8080/admin/", "")
		mediaUploadURL = flag.String("media-upload-url", "http://localhost:8080/media", "")

		webPath = flag.String("web", "web", "")
		dbPath  = flag.String("db", "file::memory:", "")
		secret  = flag.String("secret", "", "")
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

	fw, err := writer.NewFileWriter(*blogPath, *blogURL)
	if err != nil {
		log.Println(err)
		return
	}

	templates, err := blog.ParseTemplates(*webPath)
	if err != nil {
		log.Println(err)
		return
	}

	looper := &blog.Looper{}

	micropubEndpoint, mr, err := micropub.Endpoint(db, *me, looper, *mediaUploadURL, fw)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/micropub", micropubEndpoint)

	adminEndpoint, err := admin.Endpoint(*adminURL, *me, *secret, *webPath, mr, templates)
	if err != nil {
		log.Println(err)
		return
	}
	http.Handle("/admin/", http.StripPrefix("/admin", adminEndpoint))

	webmentionEndpoint, wr, err := webmention.Endpoint(db, mr, looper)
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

	looper.Blog = &blog.Blog{
		Meta: blog.Meta{
			Title:       *title,
			Description: *description,
			AuthorName:  *name,
			AuthorURL:   *me,
		},
		FileWriter: fw,
		Templates:  templates,
		Entries:    mr,
		Mentions:   wr,
	}

	serve.Serve(*port, *socket, http.DefaultServeMux)
}
