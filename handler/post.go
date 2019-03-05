package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"hawx.me/code/mux"
)

type JSON struct {
	Type       []string                 `json:"type"`
	Properties map[string][]interface{} `json:"properties"`
}

type postStore interface {
	Create(data JSON) (id string, err error)
}

func Post(store postStore) http.Handler {
	handleJSON := func(w http.ResponseWriter, r *http.Request) {
		data := JSON{Properties: map[string][]interface{}{}}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "could not decode json request: "+err.Error(), http.StatusBadRequest)
			return
		}

		id, err := store.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Location", "/"+id)
		w.WriteHeader(http.StatusCreated)
	}

	handleForm := func(w http.ResponseWriter, r *http.Request) {
		data := JSON{Properties: map[string][]interface{}{}}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		for key, values := range r.Form {
			// ignore reserved property names
			if key == "access_token" || key == "action" || key == "url" || strings.HasPrefix(key, "mp-") {
				continue
			}

			if key == "h" {
				data.Type = []string{"h-" + values[0]}
			} else {
				if strings.HasSuffix(key, "[]") {
					key := key[:len(key)-2]
					for _, value := range values {
						data.Properties[key] = append(data.Properties[key], value)
					}
				} else {
					data.Properties[key] = []interface{}{values[0]}
				}
			}
		}

		id, err := store.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Location", "/"+id)
		w.WriteHeader(http.StatusCreated)
	}

	handleMultiPart := func(w http.ResponseWriter, r *http.Request) {

	}

	return mux.ContentType{
		"application/json":                  http.HandlerFunc(handleJSON),
		"application/x-www-form-urlencoded": http.HandlerFunc(handleForm),
		"multipart/form-data":               http.HandlerFunc(handleMultiPart),
	}
}
