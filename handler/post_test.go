package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

type fakePostStore struct {
	datas []JSON
}

func (s *fakePostStore) Create(data JSON) (string, error) {
	s.datas = append(s.datas, data)
	return "1", nil
}

func TestPostEntry(t *testing.T) {
	assert := assert.New(t)
	store := &fakePostStore{}

	s := httptest.NewServer(Post(store))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"h":          {"entry"},
		"content":    {"This is a test"},
		"category[]": {"test", "ignore"},
	})

	assert.Nil(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("/1", resp.Header.Get("Location"))

	if assert.Len(store.datas, 1) {
		data := store.datas[0]

		assert.Equal("h-entry", data.Type[0])
		assert.Equal("This is a test", data.Properties["content"][0])
		assert.Equal("test", data.Properties["category"][0])
		assert.Equal("ignore", data.Properties["category"][1])
	}
}

func TestPostEntryJSON(t *testing.T) {
	assert := assert.New(t)
	store := &fakePostStore{}

	s := httptest.NewServer(Post(store))
	defer s.Close()

	resp, err := http.Post(s.URL, "application/json", strings.NewReader(`{
  "type": ["h-entry"],
  "properties": {
    "content": ["This is a test"],
    "category": ["test", "ignore"]
  }
}`))

	assert.Nil(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("/1", resp.Header.Get("Location"))

	if assert.Len(store.datas, 1) {
		data := store.datas[0]

		assert.Equal("h-entry", data.Type[0])
		assert.Equal("This is a test", data.Properties["content"][0])
		assert.Equal("test", data.Properties["category"][0])
		assert.Equal("ignore", data.Properties["category"][1])
	}
}
