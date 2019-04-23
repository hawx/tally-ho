package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/blog"
)

func Post(blog *blog.Blog) http.Handler {
	createAndRender := func(w http.ResponseWriter, data map[string][]interface{}) {
		data, err := blog.Create(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := blog.RenderPost(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		location := data["url"][0].(string)
		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusCreated)
	}

	handleSetPage := func(w http.ResponseWriter, data map[string][]interface{}) {
		if names, ok := data["name"]; !ok || len(names) != 1 {
			http.Error(w, "expected 'name'", http.StatusBadRequest)
			return
		}

		name, ok := data["name"][0].(string)
		if !ok {
			http.Error(w, "expected 'name' to be a string", http.StatusBadRequest)
			return
		}

		if err := blog.SetNextPage(name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}

	handleJSON := func(w http.ResponseWriter, r *http.Request) {
		v := jsonMicroformat{Properties: map[string][]interface{}{}}

		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, "could not decode json request: "+err.Error(), http.StatusBadRequest)
			return
		}

		data := jsonToForm(v)

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

			if err := blog.Update(v.URL, replace, add, delete); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if err := blog.Rerender(v.URL); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			return
		}

		createAndRender(w, data)
	}

	handleForm := func(w http.ResponseWriter, r *http.Request) {
		data := map[string][]interface{}{}
		setPage := false

		if err := r.ParseForm(); err != nil {
			http.Error(w, "could not parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
		for key, values := range r.Form {
			if key == "action" && len(values) == 1 && values[0] == "hx-page" {
				setPage = true
			}

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

		if setPage {
			handleSetPage(w, data)
			return
		}

		createAndRender(w, data)
	}

	handleMultiPart := func(w http.ResponseWriter, r *http.Request) {
		// todo
		log.Println("received multi-part request")
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
