package micropub

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

type fakeGetDB struct {
	entries map[string]map[string][]interface{}
}

func (b *fakeGetDB) entryByURL(url string) (map[string][]interface{}, error) {
	if entry, ok := b.entries[url]; ok {
		return entry, nil
	}

	return nil, errors.New("nope")
}

func TestConfigurationConfig(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(getHandler(nil, "http://media.example.com/"))
	defer s.Close()

	resp, err := http.Get(s.URL + "?q=config")
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	var v struct {
		MediaEndpoint string `json:"media-endpoint"`
	}
	json.NewDecoder(resp.Body).Decode(&v)
	assert.Equal("http://media.example.com/", v.MediaEndpoint)
}

func TestConfigurationSource(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeGetDB{
		entries: map[string]map[string][]interface{}{
			"https://example.com/weblog/p/1": {
				"h":     {"entry"},
				"title": {"Cool post"},
			},
		},
	}

	s := httptest.NewServer(getHandler(blog, ""))
	defer s.Close()

	resp, err := http.Get(s.URL + "?q=source&url=https://example.com/weblog/p/1")
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	var v struct {
		Type       []string
		Properties map[string][]interface{}
	}
	json.NewDecoder(resp.Body).Decode(&v)

	assert.Equal("h-entry", v.Type[0])
	assert.Equal("Cool post", v.Properties["title"][0])
}

func TestConfigurationSourceWithProperties(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeGetDB{
		entries: map[string]map[string][]interface{}{
			"https://example.com/weblog/p/1": {
				"h":     {"entry"},
				"title": {"Cool post"},
			},
		},
	}

	s := httptest.NewServer(getHandler(blog, ""))
	defer s.Close()

	resp, err := http.Get(s.URL + "?q=source&properties=title&url=https://example.com/weblog/p/1")
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	var v struct {
		Type       []string
		Properties map[string][]interface{}
	}
	json.NewDecoder(resp.Body).Decode(&v)

	assert.Len(v.Type, 0)
	assert.Equal("Cool post", v.Properties["title"][0])
}

func TestConfigurationSourceWithManyProperties(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeGetDB{
		entries: map[string]map[string][]interface{}{
			"https://example.com/weblog/p/1": {
				"h":          {"entry"},
				"title":      {"Cool post"},
				"summary":    {"goodness"},
				"categories": {"cool", "test"},
			},
		},
	}

	s := httptest.NewServer(getHandler(blog, ""))
	defer s.Close()

	resp, err := http.Get(s.URL + "?q=source&properties[]=title&properties[]=categories&url=https://example.com/weblog/p/1")
	assert.Nil(err)
	assert.Equal(http.StatusOK, resp.StatusCode)

	var v struct {
		Type       []string
		Properties map[string][]interface{}
	}
	json.NewDecoder(resp.Body).Decode(&v)

	assert.Len(v.Type, 0)
	assert.Len(v.Properties["summary"], 0)
	assert.Equal("Cool post", v.Properties["title"][0])
	assert.Equal("cool", v.Properties["categories"][0])
	assert.Equal("test", v.Properties["categories"][1])
}
