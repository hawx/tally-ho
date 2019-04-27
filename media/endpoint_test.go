package media

import (
	"bytes"
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

func (fw *fakeFileWriter) CopyToFile(path string, r io.Reader) error {
	data, _ := ioutil.ReadAll(r)
	fw.data = string(data)

	return nil
}

func (fw *fakeFileWriter) URL(path string) string {
	return "a url"
}

func (fw *fakeFileWriter) Path(url string) string {
	return "a path"
}

func TestMedia(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	fw := &fakeFileWriter{}

	s := httptest.NewServer(postHandler(fw))
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
}

func TestMediaWhenNoFilePart(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(postHandler(&fakeFileWriter{}))
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

	s := httptest.NewServer(postHandler(fw))
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
