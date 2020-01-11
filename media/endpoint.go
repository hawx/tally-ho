// Package media implements a micropub media endpoint.
//
// See the specification https://www.w3.org/TR/micropub/#media-endpoint.
package media

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"sync"

	"github.com/google/uuid"
	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/auth"
)

type FileWriter interface {
	WriteFile(name string, r io.Reader) (location string, err error)
}

type uploadState struct {
	sync.RWMutex
	LastURL string
}

// Endpoint returns a simple implementation of a media endpoint.
//
// The handler expects a multipart form with a single part named 'file'. The
// part will be written to the configured directory named with a UUID.
//
// No limits are imposed on requests made so care should be taken to configure
// this using a reverse-proxy or similar.
func Endpoint(me string, fw FileWriter) http.Handler {
	state := &uploadState{}

	return auth.Only(me, "media", mux.Method{
		"GET":  getHandler(state),
		"POST": postHandler(state, fw),
	})
}

func getHandler(state *uploadState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("q") != "last" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		state.RLock()
		lastURL := state.LastURL
		state.RUnlock()

		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(struct {
			URL string `json:"url,omitempty"`
		}{
			URL: lastURL,
		}); err != nil {
			log.Println("ERR get-last-media;", err)
		}
	}
}

func postHandler(state *uploadState, fw FileWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			log.Println("ERR media-upload;", err)
			return
		}
		if mediaType != "multipart/form-data" {
			log.Println("ERR media-upload; bad mediaType")
			http.Error(w, "expected content-type of multipart/form-data", http.StatusUnsupportedMediaType)
			return
		}

		hadPart := false
		parts := multipart.NewReader(r.Body, params["boundary"])

		part, err := parts.NextPart()
		if err == io.EOF {
			log.Println("ERR media-upload; empty form")
			http.Error(w, "expected multipart form to contain a part", http.StatusBadRequest)
			return
		}
		if err != nil {
			log.Println("ERR media-upload;", err)
			http.Error(w, "problem reading multipart form", http.StatusBadRequest)
			return
		}

		mt, ps, er := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if er != nil || mt != "form-data" || ps["name"] != "file" || hadPart {
			log.Println("ERR media-upload; expected only single part")
			http.Error(w, "request must only contain a part named 'file'", http.StatusBadRequest)
			return
		}

		uid, err := uuid.NewRandom()
		if err != nil {
			log.Println("ERR media-upload; failed to assign id")
			http.Error(w, "problem assigning id to media", http.StatusInternalServerError)
			return
		}

		ext := extension(part.Header.Get("Content-Type"), ps["filename"])

		location, err := fw.WriteFile(uid.String()+ext, part)
		if err != nil {
			log.Println("ERR media-upload;", err)
			http.Error(w, "problem writing media to file", http.StatusInternalServerError)
			return
		}

		state.Lock()
		state.LastURL = location
		state.Unlock()

		w.Header().Set("Location", location)
		w.WriteHeader(http.StatusCreated)
	}
}

func extension(contentType, filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	if len(ext) > 0 {
		return ext
	}

	exts, err := mime.ExtensionsByType(contentType)
	if err == nil && len(exts) > 0 {
		return exts[0]
	}

	return ""
}
