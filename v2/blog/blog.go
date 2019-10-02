package blog

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"hawx.me/code/route"
	"hawx.me/code/tally-ho/v2/micropub"
)

type Blog struct {
	Me          string
	Name        string
	Title       string
	Description string
	Db          *sql.DB
	Templates   *template.Template
	Posts       *micropub.Reader
}

func (b *Blog) Handler() http.Handler {
	route.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := b.Posts.Before(time.Now().UTC())
		if err != nil {
			fmt.Fprint(w, err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "list.gotmpl", posts); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/p/:id", func(w http.ResponseWriter, r *http.Request) {
		id := route.Vars(r)["id"]

		post, err := b.Posts.Post(id)
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
