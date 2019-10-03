package blog

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"hawx.me/code/route"
)

type Blog struct {
	Me          string
	Name        string
	Title       string
	Description string
	DB          *DB
	Templates   *template.Template
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

	route.HandleFunc("/p/:id", func(w http.ResponseWriter, r *http.Request) {
		post, err := b.DB.Entry(r.URL.Path)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "post.gotmpl", post); err != nil {
			fmt.Fprint(w, err)
		}
	})

	return route.Default
}
