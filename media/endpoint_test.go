package media

import (
	"bytes"
	"context"
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

func withScope(scope string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "__hawx.me/code/tally-ho:Scopes__", []string{scope})))
	})
}

func TestMedia(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}
	state := &uploadState{}

	handler := withScope("media", postHandler(state, fw))

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

	assert.Equal("a url", state.LastURL)
}

func TestMediaWithCreateScope(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}
	state := &uploadState{}

	handler := withScope("create", postHandler(state, fw))

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

	assert.Equal("a url", state.LastURL)
}

func TestMediaMissingScope(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}
	state := &uploadState{}

	handler := withScope("update", postHandler(state, fw))

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
	assert.Equal("", state.LastURL)
}

func TestMediaWhenNoFilePart(t *testing.T) {
	assert := assert.New(t)

	handler := withScope("media", postHandler(&uploadState{}, &fakeFileWriter{}))

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

	handler := withScope("media", postHandler(&uploadState{}, fw))

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

	handler := withScope("media", getHandler(&uploadState{}))

	req := httptest.NewRequest("GET", "http://localhost/?q=what", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestQueryLast(t *testing.T) {
	assert := assert.New(t)

	handler := getHandler(&uploadState{
		LastURL: "http://media.example.com/file.jpg",
	})

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

	handler := getHandler(&uploadState{})

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
