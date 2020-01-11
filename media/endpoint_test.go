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

func (fw *fakeFileWriter) WriteFile(name string, r io.Reader) (string, error) {
	data, _ := ioutil.ReadAll(r)
	fw.data = string(data)

	return "a url", nil
}

func TestMedia(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}
	state := &uploadState{}

	s := httptest.NewServer(postHandler(state, fw))
	defer s.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "whatever.png")
	assert.Nil(err)
	io.WriteString(part, file)

	assert.Nil(writer.Close())

	req, err := http.NewRequest("POST", s.URL, &buf)
	assert.Nil(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("a url", resp.Header.Get("Location"))
	assert.Equal(file, fw.data)

	assert.Equal("a url", state.LastURL)
}

func TestMediaWhenNoFilePart(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(postHandler(&uploadState{}, &fakeFileWriter{}))
	defer s.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	assert.Nil(writer.Close())

	req, err := http.NewRequest("POST", s.URL, &buf)
	assert.Nil(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestMediaWhenMultipleFileParts(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}

	s := httptest.NewServer(postHandler(&uploadState{}, fw))
	defer s.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "whatever.png")
	assert.Nil(err)
	io.WriteString(part, file+"1")

	part, err = writer.CreateFormFile("file", "other.png")
	assert.Nil(err)
	io.WriteString(part, file+"2")

	assert.Nil(writer.Close())

	req, err := http.NewRequest("POST", s.URL, &buf)
	assert.Nil(err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("a url", resp.Header.Get("Location"))
	assert.Equal(file+"1", fw.data)
}

func TestExtension(t *testing.T) {
	testCases := map[string]struct {
		ContentType, Filename, Expected string
	}{
		"from extension": {
			ContentType: "image/jpeg",
			Filename:    "file.extension",
			Expected:    ".extension",
		},
		"from uppercase extension": {
			ContentType: "image/jpeg",
			Filename:    "FILE.EXTENSION",
			Expected:    ".extension",
		},
		"from content-type": {
			ContentType: "image/jpeg",
			Filename:    "a-photo",
			Expected:    ".jpg",
		},
		"from nothing": {
			ContentType: "",
			Filename:    "",
			Expected:    "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ext := extension(tc.ContentType, tc.Filename)
			assert.Equal(t, tc.Expected, ext)
		})
	}
}

func TestQueryUnknown(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(getHandler(&uploadState{}))
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL+"?q=what", nil)
	assert.Nil(err)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestQueryLast(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(getHandler(&uploadState{
		LastURL: "http://media.example.com/file.jpg",
	}))
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL+"?q=last", nil)
	assert.Nil(err)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v map[string]string
	assert.Nil(json.NewDecoder(resp.Body).Decode(&v))
	assert.Equal("http://media.example.com/file.jpg", v["url"])
}

func TestQueryLastWhenNoneUploaded(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(getHandler(&uploadState{}))
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL+"?q=last", nil)
	assert.Nil(err)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(err)

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v map[string]string
	assert.Nil(json.NewDecoder(resp.Body).Decode(&v))
	_, ok := v["url"]
	assert.False(ok)
}
