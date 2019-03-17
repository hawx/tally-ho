package handler

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"hawx.me/code/mux"
)

type jsonMicroformat struct {
	Type       []string                 `json:"type"`
	Properties map[string][]interface{} `json:"properties"`
	Action     string                   `json:"action"`
	URL        string                   `json:"url"`
	Add        map[string][]interface{} `json:"add"`
	Delete     map[string][]interface{} `json:"delete"`
	Replace    map[string][]interface{} `json:"replace"`
}

type postStore interface {
	Create(data map[string][]interface{}) (id string, err error)
	Update(id string, replace, add, delete map[string][]interface{}) error
}

func Post(store postStore, baseURL *url.URL) http.Handler {
	handleJSON := func(w http.ResponseWriter, r *http.Request) {
		v := jsonMicroformat{Properties: map[string][]interface{}{}}

		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, "could not decode json request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(v.Type) == 0 {
			v.Type = []string{"h-entry"}
		}

		data := map[string][]interface{}{
			"h": []interface{}{
				strings.TrimPrefix(v.Type[0], "h-"),
			},
		}

		for key, value := range v.Properties {
			if reservedKey(key) {
				continue
			}

			data[key] = value
		}

		if v.Action == "update" {
			replace := map[string][]interface{}{}
			for key, value := range v.Replace {
				if reservedKey(key) {
					continue
				}

				replace[key] = value
			}

			add := map[string][]interface{}{}
			for key, value := range v.Add {
				if reservedKey(key) {
					continue
				}

				add[key] = value
			}

			delete := map[string][]interface{}{}
			for key, value := range v.Delete {
				if reservedKey(key) {
					continue
				}

				delete[key] = value
			}

			id := v.URL[len(baseURL.String())+3:]

			if err := store.Update(id, replace, add, delete); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		id, err := store.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		location, _ := baseURL.Parse("p/" + id)
		w.Header().Add("Location", location.String())
		w.WriteHeader(http.StatusCreated)
	}

	handleForm := func(w http.ResponseWriter, r *http.Request) {
		data := map[string][]interface{}{}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		for key, values := range r.Form {
			if reservedKey(key) {
				continue
			}

			if strings.HasSuffix(key, "[]") {
				key := key[:len(key)-2]
				for _, value := range values {
					data[key] = append(data[key], value)
				}
			} else {
				data[key] = []interface{}{values[0]}
			}
		}

		id, err := store.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		location, _ := baseURL.Parse("p/" + id)
		w.Header().Add("Location", location.String())
		w.WriteHeader(http.StatusCreated)
	}

	handleMultiPart := func(w http.ResponseWriter, r *http.Request) {
		// todo
	}

	return mux.ContentType{
		"application/json":                  http.HandlerFunc(handleJSON),
		"application/x-www-form-urlencoded": http.HandlerFunc(handleForm),
		"multipart/form-data":               http.HandlerFunc(handleMultiPart),
	}
}

func reservedKey(key string) bool {
	return key == "access_token" || key == "action" || key == "url" || strings.HasPrefix(key, "mp-")
}
