package micropub

import (
	"encoding/json"
	"net/http"
)

func getHandler(db *micropubDB, mediaURL string) http.HandlerFunc {
	configHandler := configHandler(mediaURL)
	sourceHandler := sourceHandler(db)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.FormValue("q") {
		case "config":
			configHandler.ServeHTTP(w, r)
		case "source":
			sourceHandler.ServeHTTP(w, r)
		}
	}
}

func configHandler(mediaURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			MediaEndpoint string `json:"media-endpoint"`
		}{
			MediaEndpoint: mediaURL,
		})
	}
}

func sourceHandler(db *micropubDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		properties := r.Form["properties[]"]
		if len(properties) == 0 {
			property := r.FormValue("properties")
			if len(property) > 0 {
				properties = []string{property}
			}
		}

		obj, err := entryByURL(db, url)
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
