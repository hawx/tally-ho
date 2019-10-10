package blog

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"hawx.me/code/route"
)

type Syndicator interface {
	Create(map[string][]interface{}) (string, error)
}

type Blog struct {
	Me          string
	Name        string
	Title       string
	Description string
	DB          *DB
	Templates   *template.Template
	Twitter     Syndicator
}

func (b *Blog) Handler() http.Handler {
	route.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := b.DB.Before(time.Now().UTC())
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "list.gotmpl", posts); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/entry/:id", func(w http.ResponseWriter, r *http.Request) {
		post, err := b.DB.Entry(r.URL.Path)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "post.gotmpl", post); err != nil {
			fmt.Fprint(w, err)
		}
	})

	// route.Handle("/:year/:month/:date/:slug")
	// route.Handle("/like/...")
	// route.Handle("/reply/...")
	// route.Handle("/tag/...")

	return route.Default
}

func (b *Blog) Entry(url string) (data map[string][]interface{}, err error) {
	return b.DB.Entry(url)
}

func (b *Blog) Create(data map[string][]interface{}) (location string, err error) {
	location, err = b.DB.Create(data)
	if err != nil {
		return
	}

	if syndicateTos, ok := data["mp-syndicate-to"]; ok && len(syndicateTos) > 0 {
		for _, syndicateTo := range syndicateTos {
			if syndicateTo == "https://twitter.com/" {
				syndicatedLocation, err := b.Twitter.Create(data)
				if err != nil {
					log.Println("syndicating to twitter: ", err)
					continue
				}

				if err := b.Update(location, empty, map[string][]interface{}{
					"syndication": {syndicatedLocation},
				}, empty); err != nil {
					log.Println("updating with twitter location: ", err)
				}
			}
		}
	}

	return
}

func (b *Blog) Update(url string, replace, add, delete map[string][]interface{}) error {
	return b.DB.Update(url, replace, add, delete)
}
