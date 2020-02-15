package blog

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

func TestGetCite(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer s.Close()

	cite, err := getCite(s.URL)
	assert.NotNil(err)
	assert.Equal(map[string]interface{}{
		"type": []string{"h-cite"},
		"properties": map[string][]interface{}{
			"url": {s.URL},
		},
	}, cite)
}

func TestGetCiteHEntry(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `<div class="h-entry">
  <div class="p-author h-card">
    <span class="p-name">John Doe</span>
  </div>
  <h1 class="p-name">A post</h1>
</div>`)
	}))
	defer s.Close()

	cite, err := getCite(s.URL)
	if !assert.Nil(err) {
		return
	}

	assert.Equal(map[string]interface{}{
		"type": []string{"h-cite"},
		"properties": map[string][]interface{}{
			"url":  {s.URL},
			"name": {"A post"},
		},
	}, cite)
}

func TestGetCiteTitle(t *testing.T) {
	assert := assert.New(t)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `<title>A post</title>`)
	}))
	defer s.Close()

	cite, err := getCite(s.URL)
	if !assert.Nil(err) {
		return
	}

	assert.Equal(map[string]interface{}{
		"type": []string{"h-cite"},
		"properties": map[string][]interface{}{
			"url":  {s.URL},
			"name": {"A post"},
		},
	}, cite)
}
