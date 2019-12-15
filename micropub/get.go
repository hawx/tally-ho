package micropub

import (
	"encoding/json"
	"net/http"

	"hawx.me/code/tally-ho/syndicate"
)

type getDB interface {
	Entry(url string) (data map[string][]interface{}, err error)
}

func getHandler(db getDB, mediaURL string, syndicators []syndicate.Syndicator) http.HandlerFunc {
	configHandler := configHandler(mediaURL, syndicators)
	sourceHandler := sourceHandler(db)
	syndicationHandler := syndicationHandler()

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.FormValue("q") {
		case "config":
			configHandler.ServeHTTP(w, r)
		case "source":
			sourceHandler.ServeHTTP(w, r)
		case "syndicate-to":
			syndicationHandler.ServeHTTP(w, r)
		}
	}
}

func configHandler(mediaURL string, syndicators []syndicate.Syndicator) http.HandlerFunc {
	configs := make([]syndicate.Config, len(syndicators))
	for i, s := range syndicators {
		configs[i] = s.Config()
	}

	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			MediaEndpoint string             `json:"media-endpoint"`
			SyndicateTo   []syndicate.Config `json:"syndicate-to"`
		}{
			MediaEndpoint: mediaURL,
			SyndicateTo:   configs,
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

		json.NewEncoder(w).Encode(formToJson(obj))
	}
}

type syndicationTarget struct {
	UID  string `json:"uid"`
	Name string `json:"name"`
}

func syndicationHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			SyndicateTo []syndicationTarget `json:"syndicate-to"`
		}{
			SyndicateTo: []syndicationTarget{
				{
					UID:  "https://twitter.com/",
					Name: "Twitter",
				},
			},
		})
	}
}
