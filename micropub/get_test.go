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

func (b *fakeGetDB) Entry(url string) (map[string][]interface{}, error) {
	if entry, ok := b.entries[url]; ok {
		return entry, nil
	}

	return nil, errors.New("nope")
}

func fakeSyndicators() []SyndicateTo {
	return []SyndicateTo{
		{UID: "https://fake/", Name: "fake on fake"},
	}
}

func TestConfigurationConfig(t *testing.T) {
	assert := assert.New(t)

	handler := getHandler(nil, "http://media.example.com/", fakeSyndicators())

	req := httptest.NewRequest("GET", "http://localhost/?q=config", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v struct {
		Q             []string `json:"q"`
		MediaEndpoint string   `json:"media-endpoint"`
		SyndicateTo   []struct {
			UID  string `json:"uid"`
			Name string `json:"name"`
		} `json:"syndicate-to"`
	}
	json.NewDecoder(resp.Body).Decode(&v)

	assert.Equal("http://media.example.com/", v.MediaEndpoint)

	assert.Equal([]string{"config", "media-endpoint", "source", "syndicate-to"}, v.Q)

	if assert.Len(v.SyndicateTo, 1) {
		assert.Equal("https://fake/", v.SyndicateTo[0].UID)
		assert.Equal("fake on fake", v.SyndicateTo[0].Name)
	}
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

	handler := getHandler(blog, "", fakeSyndicators())

	req := httptest.NewRequest("GET", "http://localhost/?q=source&url=https://example.com/weblog/p/1", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

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

	handler := getHandler(blog, "", fakeSyndicators())

	req := httptest.NewRequest("GET", "http://localhost/?q=source&properties=title&url=https://example.com/weblog/p/1", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

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

	handler := getHandler(blog, "", fakeSyndicators())

	req := httptest.NewRequest("GET", "http://localhost/?q=source&properties[]=title&properties[]=categories&url=https://example.com/weblog/p/1", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

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

func TestConfigurationSyndicationTarget(t *testing.T) {
	assert := assert.New(t)

	handler := getHandler(nil, "http://media.example.com/", fakeSyndicators())

	req := httptest.NewRequest("GET", "http://localhost/?q=syndicate-to", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v struct {
		SyndicateTo []struct {
			UID  string `json:"uid"`
			Name string `json:"name"`
		} `json:"syndicate-to"`
	}
	json.NewDecoder(resp.Body).Decode(&v)

	if assert.Len(v.SyndicateTo, 1) {
		assert.Equal("https://fake/", v.SyndicateTo[0].UID)
		assert.Equal("fake on fake", v.SyndicateTo[0].Name)
	}
}

func TestConfigurationMediaEndpoint(t *testing.T) {
	assert := assert.New(t)

	handler := getHandler(nil, "http://media.example.com/", fakeSyndicators())

	req := httptest.NewRequest("GET", "http://localhost/?q=media-endpoint", nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()

	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Header.Get("Content-Type"))

	var v struct {
		MediaEndpoint string `json:"media-endpoint"`
	}
	json.NewDecoder(resp.Body).Decode(&v)

	assert.Equal("http://media.example.com/", v.MediaEndpoint)
}
