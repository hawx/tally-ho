package micropub

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"hawx.me/code/mux"
)

type blogInterface interface {
	Create(data map[string][]interface{}) (map[string][]interface{}, error)
	Update(url string, replace, add, delete map[string][]interface{}) error
	SetNextPage(name string) error

	// these should be removed, blog should know what/when to render
	Rerender(url string) error
	RenderPost(data map[string][]interface{}) error
}

func postHandler(blog blogInterface) http.Handler {
	h := micropubPostHandler{
		blog: blog,
	}

	return mux.ContentType{
		"application/json":                  http.HandlerFunc(h.handleJSON),
		"application/x-www-form-urlencoded": http.HandlerFunc(h.handleForm),
		"multipart/form-data":               http.HandlerFunc(h.handleMultiPart),
	}
}

type micropubPostHandler struct {
	blog blogInterface
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

		if err := h.blog.Update(v.URL, replace, add, delete); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := h.blog.Rerender(v.URL); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.createAndRender(w, data)
}

func (h *micropubPostHandler) handleForm(w http.ResponseWriter, r *http.Request) {
	data := map[string][]interface{}{}
	isSetPage := false

	if err := r.ParseForm(); err != nil {
		http.Error(w, "could not parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	for key, values := range r.Form {
		if key == "action" && len(values) == 1 && values[0] == "hx-page" {
			isSetPage = true
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

	if isSetPage {
		h.setPage(w, data)
		return
	}

	h.createAndRender(w, data)
}

func (h *micropubPostHandler) handleMultiPart(w http.ResponseWriter, r *http.Request) {
	// todo
	log.Println("received multi-part request")
}

func (h *micropubPostHandler) createAndRender(w http.ResponseWriter, data map[string][]interface{}) {
	data, err := h.blog.Create(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.blog.RenderPost(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	location := data["url"][0].(string)
	w.Header().Add("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *micropubPostHandler) setPage(w http.ResponseWriter, data map[string][]interface{}) {
	if names, ok := data["name"]; !ok || len(names) != 1 {
		http.Error(w, "expected 'name'", http.StatusBadRequest)
		return
	}

	name, ok := data["name"][0].(string)
	if !ok {
		http.Error(w, "expected 'name' to be a string", http.StatusBadRequest)
		return
	}

	if err := h.blog.SetNextPage(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func reservedKey(key string) bool {
	return key == "access_token" || key == "action" || key == "url" || strings.HasPrefix(key, "mp-")
}
