package micropub

import (
	"encoding/json"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/auth"
	"hawx.me/code/tally-ho/media"
)

type postDB interface {
	Create(data map[string][]any) (string, error)
	Update(url string, replace, add, delete map[string][]any, deleteAlls []string) error
	Delete(url string) error
	Undelete(url string) error
}

func postHandler(db postDB, fw media.FileWriter) http.Handler {
	h := micropubPostHandler{
		db: db,
		fw: fw,
	}

	return mux.ContentType{
		"application/json":                  http.HandlerFunc(h.handleJSON),
		"application/x-www-form-urlencoded": http.HandlerFunc(h.handleForm),
		"multipart/form-data":               http.HandlerFunc(h.handleMultiPart),
	}
}

type micropubPostHandler struct {
	db postDB
	fw media.FileWriter
}

func (h *micropubPostHandler) handleJSON(w http.ResponseWriter, r *http.Request) {
	v := jsonMicroformat{Properties: map[string][]any{}}

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		http.Error(w, "could not decode json request: "+err.Error(), http.StatusBadRequest)
		return
	}

	data := jsonToForm(v)

	if v.Action == "update" {
		replace := map[string][]any{}
		for key, value := range v.Replace {
			if reservedKey(key) {
				continue
			}

			replace[key] = value
		}

		add := map[string][]any{}
		for key, value := range v.Add {
			if reservedKey(key) {
				continue
			}

			add[key] = value
		}

		delete := map[string][]any{}
		var deleteAlls []string

		if ds, ok := v.Delete.([]any); ok {
			for _, d := range ds {
				if dd, ok := d.(string); ok {
					deleteAlls = append(deleteAlls, dd)
				} else {
					http.Error(w, "could not decode json request: malformed delete", http.StatusBadRequest)
					return
				}
			}
		} else if dm, ok := v.Delete.(map[string]any); ok {
			for key, value := range dm {
				if reservedKey(key) {
					continue
				}

				if vs, ok := value.([]any); ok {
					delete[key] = vs
				} else {
					http.Error(w, "could not decode json request: malformed delete", http.StatusBadRequest)
					return
				}
			}
		}

		if !auth.HasScope(w, r, "update") {
			return
		}

		if err := h.db.Update(v.URL, replace, add, delete, deleteAlls); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	if v.Action == "delete" {
		h.delete(w, r, v.URL)
		return
	}

	if v.Action == "undelete" {
		h.undelete(w, r, v.URL)
		return
	}

	h.create(w, r, data)
}

func (h *micropubPostHandler) handleForm(w http.ResponseWriter, r *http.Request) {
	data := map[string][]any{}

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
				if value != "" {
					data[key] = append(data[key], value)
				}
			}
		} else {
			if values[0] != "" {
				data[key] = []any{values[0]}
			}
		}
	}

	if r.FormValue("action") == "delete" {
		h.delete(w, r, r.FormValue("url"))
		return
	}

	if r.FormValue("action") == "undelete" {
		h.undelete(w, r, r.FormValue("url"))
		return
	}

	h.create(w, r, data)
}

func (h *micropubPostHandler) handleMultiPart(w http.ResponseWriter, r *http.Request) {
	if !auth.HasScope(w, r, "create") {
		return
	}

	_, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		slog.Error("micropub parse content type", slog.Any("err", err))
		return
	}

	data := map[string][]any{}
	parts := multipart.NewReader(r.Body, params["boundary"])

	for {
		p, err := parts.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			slog.Error("micropub read multi part", slog.Any("err", err))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		mt, ps, er := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
		if er != nil || mt != "form-data" {
			continue
		}

		key := p.FormName()

		switch key {
		case "photo", "video", "audio":
			location, err := h.fw.WriteFile(ps["filename"], p.Header.Get("Content-Type"), p)
			if err != nil {
				slog.Error("micropub photo", slog.Any("err", err))
				continue
			}

			data[key] = []any{location}

		case "photo[]", "video[]", "audio[]":
			location, err := h.fw.WriteFile(ps["filename"], p.Header.Get("Content-Type"), p)
			if err != nil {
				slog.Error("micropub-photo", slog.Any("err", err))
				continue
			}

			key := key[:len(key)-2]
			data[key] = append(data[key], location)

		default:
			if reservedKey(key) {
				continue
			}

			slurp, err := io.ReadAll(p)
			if err != nil {
				slog.Error("could not read", slog.Any("err", err))
				return
			}

			value := string(slurp)
			if value == "" {
				continue
			}

			if strings.HasSuffix(key, "[]") {
				key := key[:len(key)-2]
				data[key] = append(data[key], value)
			} else {
				data[key] = []any{value}
			}
		}
	}

	h.create(w, r, data)
}

func (h *micropubPostHandler) create(w http.ResponseWriter, r *http.Request, data map[string][]any) {
	if !auth.HasScope(w, r, "create") {
		return
	}

	if clientID := auth.ClientID(r); clientID != "" {
		data["hx-client-id"] = []any{clientID}
	}

	location, err := h.db.Create(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *micropubPostHandler) delete(w http.ResponseWriter, r *http.Request, url string) {
	if !auth.HasScope(w, r, "delete") {
		return
	}

	if err := h.db.Delete(url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *micropubPostHandler) undelete(w http.ResponseWriter, r *http.Request, url string) {
	if !auth.HasScope(w, r, "delete") {
		return
	}

	if err := h.db.Undelete(url); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func reservedKey(key string) bool {
	return key == "access_token" || key == "action" || key == "url"
}
