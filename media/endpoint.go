package media

import (
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"

	"github.com/google/uuid"
	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/blog"
)

// Endpoint returns a simple implementation of a media endpoint. Files will be
// written to 'path' and then are expected to be hosted from 'url'.
//
//   Endpoint("/wwwfiles/media/", "https://example.com/media/")
//
// The handler expects a multipart form with at least a single part named
// 'file'. The first such part will be written to the configured directory named
// with a uuid, any additional parts will be ignored.
//
// No limits are imposed on requests made so care should be taken to configure
// this using a reverse-proxy or similar.
func Endpoint(path, url string) (h http.Handler, err error) {
	fw, err := blog.NewFileWriter(path, url)
	if err != nil {
		return
	}

	return mux.Method{"POST": postHandler(fw)}, nil
}

func postHandler(fw blog.FileWriter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil {
			log.Fatal(err)
		}
		if mediaType != "multipart/form-data" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		hadPart := false
		parts := multipart.NewReader(r.Body, params["boundary"])

		part, err := parts.NextPart()
		if err == io.EOF {
			http.Error(w, "expected multipart form to contain a part", http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, "problem reading multipart form", http.StatusBadRequest)
			return
		}

		mt, ps, er := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
		if er != nil || mt != "form-data" || ps["name"] != "file" || hadPart {
			http.Error(w, "request must only contain a part named 'file'", http.StatusBadRequest)
			return
		}

		uid, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, "problem assigning id to media", http.StatusInternalServerError)
			return
		}

		if err := fw.CopyToFile(uid.String(), part); err != nil {
			http.Error(w, "problem writing media to file", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", fw.URL(uid.String()))
		w.WriteHeader(http.StatusCreated)
	}
}
