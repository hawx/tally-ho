package main

import (
	"flag"
	"log"
	"net/http"

	"hawx.me/code/indieauth"
	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/handler"
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

	mediaWriter, err := blog.NewFileWriter(*mediaPath, *mediaURL)
	if err != nil {
		log.Println("creating mediawriter:", err)
		return
	}

	blog, err := blog.New(blog.Options{
		WebPath:  *webPath,
		BaseURL:  *baseURL,
		BasePath: *basePath,
		DbPath:   *dbPath,
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer blog.Close()

	if flag.NArg() == 1 && flag.Arg(0) == "render" {
		if err := blog.RenderAll(); err != nil {
			log.Fatal(err)
		}

		return
	}

	auth, err := indieauth.Authorization(*adminURL, *adminURL+"callback", []string{"create"})
	if err != nil {
		log.Fatal(err)
	}

	session, err := handler.NewScopedSessions(*me, *secret, auth)
	if err != nil {
		log.Fatal(err)
	}
	session.Root = *adminURL

	if *me == "" {
		log.Fatal("--me must be provided")
	}

	route.HandleFunc("/admin/sign-in", session.SignIn())
	route.HandleFunc("/admin/callback", session.Callback())
	route.HandleFunc("/admin/sign-out", session.SignOut())

	route.Handle("/admin", mux.Method{
		"GET": session.WithToken(handler.Admin(blog, *adminURL)),
	})
	route.Handle("/admin/public/*path", http.StripPrefix("/admin/public", http.FileServer(http.Dir(*webPath+"/static"))))

	route.Handle("/micropub", handler.Authenticate(*me, "create", mux.Method{
		"POST": handler.Post(blog),
		"GET":  handler.Configuration(blog, *mediaUploadURL),
	}))

	route.Handle("/webmention", mux.Method{
		"POST": handler.Mention(blog),
	})

	route.Handle("/media", mux.Method{
		"POST": handler.Media(mediaWriter),
	})

	route.Handle("/public/*path", http.StripPrefix("/public", http.FileServer(http.Dir(*webPath+"/static"))))

	serve.Serve(*port, *socket, route.Default)
}
