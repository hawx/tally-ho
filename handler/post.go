package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"hawx.me/code/mux"
)

type jsonMicroformat struct {
	Type       []string                 `json:"type"`
	Properties map[string][]interface{} `json:"properties"`
}

type postStore interface {
	Create(data map[string][]interface{}) (id string, err error)
}

func Post(store postStore) http.Handler {
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

		id, err := store.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Location", "/"+id)
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

		w.Header().Add("Location", "/"+id)
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
