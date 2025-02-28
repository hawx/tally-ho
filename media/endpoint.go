// Package media implements a micropub media endpoint.
//
// See the specification https://www.w3.org/TR/micropub/#media-endpoint.
package media

import (
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"sync"
)

type FileWriter interface {
	// WriteFile is given the name and content-type of the uploaded media, along
	// with a Reader of its contents. It returns a URL locating the file, or an
	// error if a problem occured.
	WriteFile(name, contentType string, r io.Reader) (location string, err error)
}

// HasScope returns true if the Request contains one of the listed valid
// scopes. It is expected to write any applicable error information and/or
// status codes to the ResponseWriter.
type HasScope func(w http.ResponseWriter, r *http.Request, valid ...string) bool

type Handler struct {
	logger   *slog.Logger
	fw       FileWriter
	hasScope HasScope

	mu      sync.RWMutex
	lastURL string
}

// Endpoint returns a simple implementation of a media endpoint. It expects a
// multipart form with a single part named 'file'.
//
// No limits are imposed on requests so care should be taken to configure them
// using a reverse-proxy or similar.
//
// The URL of the last file uploaded can be queried by requesting 'GET
// /?q=last'.
func Endpoint(fw FileWriter, hasScope HasScope) *Handler {
	return &Handler{
		logger:   slog.Default().With("component", "media"),
		fw:       fw,
		hasScope: hasScope,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPost:
		h.post(w, r)
	case http.MethodOptions:
		w.Header().Set("Accept", "GET,POST")
	default:
		w.Header().Set("Accept", "GET,POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("q") != "last" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	lastURL := h.lastURL
	h.mu.RUnlock()

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(struct {
		URL string `json:"url,omitempty"`
	}{
		URL: lastURL,
	}); err != nil {
		h.logger.Error("get last media", slog.Any("err", err))
	}
}

func (h *Handler) post(w http.ResponseWriter, r *http.Request) {
	if !h.hasScope(w, r, "media", "create") {
		return
	}

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error("parsing media type", slog.Any("err", err))
		return
	}
	if mediaType != "multipart/form-data" {
		h.logger.Error("bad mediaType")
		http.Error(w, "expected content-type of multipart/form-data", http.StatusUnsupportedMediaType)
		return
	}

	parts := multipart.NewReader(r.Body, params["boundary"])

	part, err := parts.NextPart()
	if err == io.EOF {
		h.logger.Error("empty form")
		http.Error(w, "expected multipart form to contain a part", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.logger.Error("next part", slog.Any("err", err))
		http.Error(w, "problem reading multipart form", http.StatusBadRequest)
		return
	}

	mt, ps, er := mime.ParseMediaType(part.Header.Get("Content-Disposition"))
	if er != nil || mt != "form-data" || ps["name"] != "file" {
		h.logger.Error("expected only single part")
		http.Error(w, "request must only contain a part named 'file'", http.StatusBadRequest)
		return
	}

	location, err := h.fw.WriteFile(ps["filename"], part.Header.Get("Content-Type"), part)
	if err != nil {
		h.logger.Error("write file", slog.Any("err", err))
		http.Error(w, "problem writing media to file", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.lastURL = location
	h.mu.Unlock()

	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}
