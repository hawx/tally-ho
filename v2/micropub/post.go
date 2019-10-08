package micropub

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"hawx.me/code/mux"
)

type postDB interface {
	Create(data map[string][]interface{}) (string, error)
	Update(url string, replace, add, delete map[string][]interface{}) error
}

func postHandler(db postDB) http.Handler {
	h := micropubPostHandler{
		db: db,
	}

	return mux.ContentType{
		"application/json":                  http.HandlerFunc(h.handleJSON),
		"application/x-www-form-urlencoded": http.HandlerFunc(h.handleForm),
		"multipart/form-data":               http.HandlerFunc(h.handleMultiPart),
	}
}

type micropubPostHandler struct {
	db postDB
}

func (h *micropubPostHandler) handleJSON(w http.ResponseWriter, r *http.Request) {
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

		if err := h.db.Update(v.URL, replace, add, delete); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.create(w, data)
}

func (h *micropubPostHandler) handleForm(w http.ResponseWriter, r *http.Request) {
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

	h.create(w, data)
}

func (h *micropubPostHandler) handleMultiPart(w http.ResponseWriter, r *http.Request) {
	// todo
	log.Println("received multi-part request")
}

func (h *micropubPostHandler) create(w http.ResponseWriter, data map[string][]interface{}) {
	location, err := h.db.Create(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func reservedKey(key string) bool {
	return key == "access_token" || key == "action" || key == "url" || strings.HasPrefix(key, "mp-")
}