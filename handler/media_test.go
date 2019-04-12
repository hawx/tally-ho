package handler

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

type fakeMediaBlog struct {
	data string
}

func (b *fakeMediaBlog) WriteMedia(r io.Reader) (string, error) {
	data, _ := ioutil.ReadAll(r)
	b.data = string(data)

	return "http://some.location/uuid", nil
}

func TestMedia(t *testing.T) {
	assert := assert.New(t)
	file := "this is an image"
	blog := &fakeMediaBlog{}

	s := httptest.NewServer(Media(blog))
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
	assert.Equal("http://some.location/uuid", resp.Header.Get("Location"))
	assert.Equal(file, blog.data)
}

func TestMediaWhenNoFilePart(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(Media(&fakeMediaBlog{}))
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
	blog := &fakeMediaBlog{}

	s := httptest.NewServer(Media(blog))
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
	assert.Equal("http://some.location/uuid", resp.Header.Get("Location"))
	assert.Equal(file+"1", blog.data)
}

// func TestMediaWhenFilePartIsTooLarge(t *testing.T) {
// 	assert := assert.New(t)
// 	file := strings.NewReader(strings.Repeat("a", 50*1024*1024))

// 	s := httptest.NewServer(Media(&fakeMediaBlog{}))
// 	defer s.Close()

// 	var buf bytes.Buffer
// 	writer := multipart.NewWriter(&buf)

// 	part, err := writer.CreateFormFile("file", "whatever.png")
// 	assert.Nil(err)
// 	io.Copy(part, file)

// 	assert.Nil(writer.Close())

// 	req, err := http.NewRequest("POST", s.URL, &buf)
// 	assert.Nil(err)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	resp, err := http.DefaultClient.Do(req)
// 	assert.Nil(err)

// 	assert.Equal(http.StatusCreated, resp.StatusCode)
// 	assert.Equal("http://some.location/uuid", resp.Header.Get("Location"))
// }
