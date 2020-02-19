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
		"type": []interface{}{"h-cite"},
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
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"url":  {s.URL},
			"name": {"A post"},
			"author": {
				map[string]interface{}{
					"type": []interface{}{"h-card"},
					"properties": map[string][]interface{}{
						"name": {"John Doe"},
					},
				},
			},
		},
	}, cite)
}

func TestGetCiteOG(t *testing.T) {
	testCases := map[string]struct {
		html     string
		err      error
		expected map[string]interface{}
	}{
		"not-article": {
			html: `<meta property="og:type" content="post" />
<meta property="og:title" content="A post" />`,
			err: ErrNoName,
			expected: map[string]interface{}{
				"type":       []interface{}{"h-cite"},
				"properties": map[string][]interface{}{},
			},
		},
		"only-title": {
			html: `<meta property="og:type" content="article" />
<meta property="og:title" content="A post" />`,
			expected: map[string]interface{}{
				"type": []interface{}{"h-cite"},
				"properties": map[string][]interface{}{
					"name": {"A post"},
				},
			},
		},
		"no-title": {
			html: `<meta property="og:type" content="article" />
<meta property="og:site_name" content="John's Blog" />
<meta property="og:url" content="https://example.com/no-this-one" />`,
			err: ErrNoName,
			expected: map[string]interface{}{
				"type":       []interface{}{"h-cite"},
				"properties": map[string][]interface{}{},
			},
		},
		"all": {
			html: `<meta property="og:type" content="article" />
<meta property="og:site_name" content="John's Blog" />
<meta property="og:title" content="A post" />
<meta property="og:url" content="https://example.com/no-this-one" />`,
			expected: map[string]interface{}{
				"type": []interface{}{"h-cite"},
				"properties": map[string][]interface{}{
					"url":  {"https://example.com/no-this-one"},
					"name": {"A post"},
					"author": {
						map[string]interface{}{
							"type": []interface{}{"h-card"},
							"properties": map[string][]interface{}{
								"name": {"John's Blog"},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, tc.html)
			}))
			defer s.Close()

			cite, err := getCite(s.URL)
			if !assert.Equal(tc.err, err) {
				return
			}

			if _, ok := tc.expected["properties"].(map[string][]interface{})["url"]; !ok {
				tc.expected["properties"].(map[string][]interface{})["url"] = []interface{}{s.URL}
			}

			assert.Equal(tc.expected, cite)
		})
	}
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
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"url":  {s.URL},
			"name": {"A post"},
		},
	}, cite)
}
