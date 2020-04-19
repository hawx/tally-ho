package media

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

type fakeFileWriter struct {
	data string
}

func (fw *fakeFileWriter) WriteFile(name, contentType string, r io.Reader) (string, error) {
	data, _ := ioutil.ReadAll(r)
	fw.data = string(data)

	return "a url", nil
}

type fakeHasScope struct {
	ok    bool
	valid []string
}

func (hs *fakeHasScope) HasScope(w http.ResponseWriter, r *http.Request, valid ...string) bool {
	hs.valid = valid

	if !hs.ok {
		w.WriteHeader(http.StatusUnauthorized)
	}

	return hs.ok
}

func hasScope(ok bool) HasScope {
	a := &fakeHasScope{ok: ok}
	return a.HasScope
}

func TestMedia(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}
	hs := &fakeHasScope{ok: true}
	state := Endpoint(fw, hs.HasScope)

	handler := state

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "whatever.png")
	assert.Nil(err)
	io.WriteString(part, file)

	assert.Nil(writer.Close())

	req := httptest.NewRequest("POST", "http://localhost/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("a url", resp.Header.Get("Location"))
	assert.Equal(file, fw.data)

	assert.Equal("a url", state.lastURL)
	assert.Equal([]string{"media", "create"}, hs.valid)
}

func TestMediaMissingScope(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}
	state := Endpoint(fw, hasScope(false))

	handler := state

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "whatever.png")
	assert.Nil(err)
	io.WriteString(part, file)

	assert.Nil(writer.Close())

	req := httptest.NewRequest("POST", "http://localhost/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusUnauthorized, resp.StatusCode)
	assert.Equal("", fw.data)
	assert.Equal("", state.lastURL)
}

func TestMediaWhenNoFilePart(t *testing.T) {
	assert := assert.New(t)

	state := Endpoint(&fakeFileWriter{}, hasScope(true))
	handler := state

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	assert.Nil(writer.Close())

	req := httptest.NewRequest("POST", "http://localhost/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestMediaWhenMultipleFileParts(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}

	state := Endpoint(fw, hasScope(true))
	handler := state

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "whatever.png")
	assert.Nil(err)
	io.WriteString(part, file+"1")

	part, err = writer.CreateFormFile("file", "other.png")
	assert.Nil(err)
	io.WriteString(part, file+"2")

	assert.Nil(writer.Close())

	req := httptest.NewRequest("POST", "http://localhost/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("a url", resp.Header.Get("Location"))
	assert.Equal(file+"1", fw.data)
}

func TestQueryUnknown(t *testing.T) {
	assert := assert.New(t)

	state := Endpoint(nil, hasScope(true))
	handler := state

	req := httptest.NewRequest("GET", "http://localhost/?q=what", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestQueryLast(t *testing.T) {
	assert := assert.New(t)

	state := Endpoint(nil, hasScope(true))
	state.lastURL = "http://media.example.com/file.jpg"
	handler := state

	req := httptest.NewRequest("GET", "http://localhost/?q=last", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v map[string]string
	assert.Nil(json.NewDecoder(resp.Body).Decode(&v))
	assert.Equal("http://media.example.com/file.jpg", v["url"])
}

func TestQueryLastWhenNoneUploaded(t *testing.T) {
	assert := assert.New(t)

	handler := Endpoint(nil, hasScope(true))

	req := httptest.NewRequest("GET", "http://localhost/?q=last", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v map[string]string
	assert.Nil(json.NewDecoder(resp.Body).Decode(&v))
	_, ok := v["url"]
	assert.False(ok)
}
