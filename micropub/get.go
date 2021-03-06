package micropub

import (
	"encoding/json"
	"net/http"
)

type getDB interface {
	Entry(url string) (data map[string][]interface{}, err error)
}

func getHandler(
	db getDB,
	mediaURL string,
	syndicateTo []SyndicateTo,
) http.HandlerFunc {
	configHandler := configHandler(mediaURL, syndicateTo)
	sourceHandler := sourceHandler(db)
	syndicationHandler := syndicationHandler(syndicateTo)
	mediaEndpointHandler := mediaEndpointHandler(mediaURL)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.FormValue("q") {
		case "config":
			configHandler.ServeHTTP(w, r)
		case "media-endpoint":
			mediaEndpointHandler.ServeHTTP(w, r)
		case "source":
			sourceHandler.ServeHTTP(w, r)
		case "syndicate-to":
			syndicationHandler.ServeHTTP(w, r)
		}
	}
}

type SyndicateTo struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func configHandler(mediaURL string, syndicateTo []SyndicateTo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Q             []string      `json:"q"`
			MediaEndpoint string        `json:"media-endpoint"`
			SyndicateTo   []SyndicateTo `json:"syndicate-to"`
		}{
			Q: []string{
				"config",
				"media-endpoint",
				"source",
				"syndicate-to",
			},
			MediaEndpoint: mediaURL,
			SyndicateTo:   syndicateTo,
		})
	}
}

func mediaEndpointHandler(mediaURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			MediaEndpoint string `json:"media-endpoint"`
		}{
			MediaEndpoint: mediaURL,
		})
	}
}

func sourceHandler(db getDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := r.FormValue("url")
		properties := r.Form["properties[]"]
		if len(properties) == 0 {
			property := r.FormValue("properties")
			if len(property) > 0 {
				properties = []string{property}
			}
		}

		obj, err := db.Entry(url)
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(formToJSON(obj))
	}
}

type syndicationTarget struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func syndicationHandler(syndicateTo []SyndicateTo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			SyndicateTo []SyndicateTo `json:"syndicate-to"`
		}{
			SyndicateTo: syndicateTo,
		})
	}
}

func contains(needle string, list []string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}

	return false
}
