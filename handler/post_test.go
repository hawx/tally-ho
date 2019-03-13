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
	datas                   []map[string][]interface{}
	replaces, adds, deletes map[string][]map[string][]interface{}
}

func (s *fakePostStore) Create(data map[string][]interface{}) (string, error) {
	s.datas = append(s.datas, data)
	return "1", nil
}

func (s *fakePostStore) Update(url string, replace, add, delete map[string][]interface{}) error {
	s.replaces[url] = append(s.replaces[url], replace)
	s.adds[url] = append(s.adds[url], add)
	s.deletes[url] = append(s.deletes[url], delete)

	return nil
}

func TestPostEntry(t *testing.T) {
	assert := assert.New(t)
	store := &fakePostStore{}

	s := httptest.NewServer(Post(store))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"h":            {"entry"},
		"content":      {"This is a test"},
		"category[]":   {"test", "ignore"},
		"mp-something": {"what"},
	})

	assert.Nil(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("/1", resp.Header.Get("Location"))

	if assert.Len(store.datas, 1) {
		data := store.datas[0]

		assert.Equal("entry", data["h"][0])
		assert.Equal("This is a test", data["content"][0])
		assert.Equal("test", data["category"][0])
		assert.Equal("ignore", data["category"][1])

		_, ok := data["mp-something"]
		assert.False(ok)
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
    "category": ["test", "ignore"],
    "mp-something": ["what"]
  }
}`))

	assert.Nil(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("/1", resp.Header.Get("Location"))

	if assert.Len(store.datas, 1) {
		data := store.datas[0]

		assert.Equal("entry", data["h"][0])
		assert.Equal("This is a test", data["content"][0])
		assert.Equal("test", data["category"][0])
		assert.Equal("ignore", data["category"][1])

		_, ok := data["mp-something"]
		assert.False(ok)
	}
}

func TestUpdateEntry(t *testing.T) {
	assert := assert.New(t)
	store := &fakePostStore{
		adds:     map[string][]map[string][]interface{}{},
		deletes:  map[string][]map[string][]interface{}{},
		replaces: map[string][]map[string][]interface{}{},
	}

	s := httptest.NewServer(Post(store))
	defer s.Close()

	resp, err := http.Post(s.URL, "application/json", strings.NewReader(`{
  "action": "update",
  "url": "https://example.com/post/100",
  "replace": {
    "content": ["hello moon"]
  },
  "add": {
    "syndication": ["http://somewhere.com"]
  },
  "delete": {
    "not-important": ["this"]
  }
}`))

	assert.Nil(err)
	assert.Equal(http.StatusNoContent, resp.StatusCode)

	replace, ok := store.replaces["https://example.com/post/100"]
	if assert.True(ok) && assert.Len(replace, 1) {
		assert.Equal("hello moon", replace[0]["content"][0])
	}

	add, ok := store.adds["https://example.com/post/100"]
	if assert.True(ok) && assert.Len(add, 1) {
		assert.Equal("http://somewhere.com", add[0]["syndication"][0])
	}

	delete, ok := store.deletes["https://example.com/post/100"]
	if assert.True(ok) && assert.Len(delete, 1) {
		assert.Equal("this", delete[0]["not-important"][0])
	}
}
