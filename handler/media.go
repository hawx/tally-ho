package handler

import (
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
)

type mediaBlog interface {
	WriteMedia(r io.Reader) (location string, err error)
}

func Media(blog mediaBlog) http.HandlerFunc {
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
		for {
			part, err := parts.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatal(err)
			}
			mt, ps, er := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
			if er != nil || mt != "form-data" || ps["name"] != "file" || hadPart {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hadPart = true
			location, err := blog.WriteMedia(part)

			w.Header().Set("Location", location)
			w.WriteHeader(http.StatusCreated)
			return
		}

		if !hadPart {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}
