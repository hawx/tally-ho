package handler

import (
	"encoding/json"
	"net/http"

	"hawx.me/code/tally-ho/blog"
)

func Configuration(blog *blog.Blog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("q") == "config" {
			w.Write([]byte("{}")) // for now
		}

		if r.FormValue("q") == "source" {
			url := r.FormValue("url")
			properties := r.Form["properties[]"]
			if len(properties) == 0 {
				property := r.FormValue("properties")
				if len(property) > 0 {
					properties = []string{property}
				}
			}

			obj, err := blog.PostByURL(url)
			if err != nil {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			if len(properties) > 0 {
				for key := range obj {
					if !contains(key, properties) {
						delete(obj, key)
					}
				}
			}

			json.NewEncoder(w).Encode(formToJson(obj))
		}
	}
}
