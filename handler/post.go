package handler

import (
	"io"
	"log"
	"net/http"
	"os"
)

type postStore interface {
	CreateEntry(content string, categories []string) error
}

func Post(store postStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h := r.FormValue("h")

		switch h {
		case "entry":
			content := r.FormValue("content")

			log.Println("new post\n\n", content)
		default:
			log.Println("don't know how to ", h)
			io.Copy(os.Stdout, r.Body)
		}
	}
}
