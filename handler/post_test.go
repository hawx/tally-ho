package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"hawx.me/code/assert"
)

type fakeEntry struct {
	content    string
	categories []string
}

type fakePostStore struct {
	entries []fakeEntry
}

func (s *fakePostStore) CreateEntry(content string, categories []string) error {
	s.entries = append(s.entries, fakeEntry{content, categories})
	return nil
}

func TestPostEntry(t *testing.T) {
	assert := assert.New(t)
	store := &fakePostStore{}

	s := httptest.NewServer(Post(store))
	defer s.Close()

	http.PostForm(s.URL, url.Values{
		"h":          {"entry"},
		"content":    {"This is a test"},
		"category[]": {"test", "ignore"},
	})

	assert.Len(store.entries, 1)
}

func TestPostEntryJSON(t *testing.T) {
	t.Fatal("todo")
}
