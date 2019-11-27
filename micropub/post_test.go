package micropub

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"hawx.me/code/assert"
	"hawx.me/code/tally-ho/writer"
)

type fakePostDB struct {
	datas                   []map[string][]interface{}
	replaces, adds, deletes map[string][]map[string][]interface{}
}

func (b *fakePostDB) Create(data map[string][]interface{}) (string, error) {
	b.datas = append(b.datas, data)

	return "http://example.com/blog/p/1", nil
}

func (b *fakePostDB) Update(id string, replace, add, delete map[string][]interface{}) error {
	b.replaces[id] = append(b.replaces[id], replace)
	b.adds[id] = append(b.adds[id], add)
	b.deletes[id] = append(b.deletes[id], delete)

	return nil
}

func (b *fakePostDB) setNextPage(uf writer.URLFactory, name string) error {
	return nil
}

func TestPostEntry(t *testing.T) {
	assert := assert.New(t)
	blog := &fakePostDB{}

	s := httptest.NewServer(postHandler(blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"h":            {"entry"},
		"content":      {"This is a test"},
		"category[]":   {"test", "ignore"},
		"mp-something": {"what"},
		"url":          {"what"},
	})

	assert.Nil(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("http://example.com/blog/p/1", resp.Header.Get("Location"))

	if assert.Len(blog.datas, 1) {
		data := blog.datas[0]

		assert.Equal("entry", data["h"][0])
		assert.Equal("This is a test", data["content"][0])
		assert.Equal("test", data["category"][0])
		assert.Equal("ignore", data["category"][1])
		assert.Equal("what", data["mp-something"][0])

		_, ok := data["url"]
		assert.False(ok)
	}
}

func TestPostEntryJSON(t *testing.T) {
	assert := assert.New(t)
	db := &fakePostDB{}

	s := httptest.NewServer(postHandler(db))
	defer s.Close()

	resp, err := http.Post(s.URL, "application/json", strings.NewReader(`{
  "type": ["h-entry"],
  "properties": {
    "content": ["This is a test"],
    "category": ["test", "ignore"],
    "mp-something": ["what"],
    "url": ["http://what"]
  }
}`))

	assert.Nil(err)
	assert.Equal(http.StatusCreated, resp.StatusCode)
	assert.Equal("http://example.com/blog/p/1", resp.Header.Get("Location"))

	if assert.Len(db.datas, 1) {
		data := db.datas[0]

		assert.Equal("entry", data["h"][0])
		assert.Equal("This is a test", data["content"][0])
		assert.Equal("test", data["category"][0])
		assert.Equal("ignore", data["category"][1])
		assert.Equal("what", data["mp-something"][0])

		_, ok := data["url"]
		assert.False(ok)
	}
}

func TestUpdateEntry(t *testing.T) {
	assert := assert.New(t)
	db := &fakePostDB{
		adds:     map[string][]map[string][]interface{}{},
		deletes:  map[string][]map[string][]interface{}{},
		replaces: map[string][]map[string][]interface{}{},
	}

	s := httptest.NewServer(postHandler(db))
	defer s.Close()

	resp, err := http.Post(s.URL, "application/json", strings.NewReader(`{
  "action": "update",
  "url": "https://example.com/blog/p/100",
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

	replace, ok := db.replaces["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(replace, 1) {
		assert.Equal("hello moon", replace[0]["content"][0])
	}

	add, ok := db.adds["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(add, 1) {
		assert.Equal("http://somewhere.com", add[0]["syndication"][0])
	}

	delete, ok := db.deletes["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(delete, 1) {
		assert.Equal("this", delete[0]["not-important"][0])
	}
}
