package micropub

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

type fakePostBlog struct {
	datas                   []map[string][]interface{}
	replaces, adds, deletes map[string][]map[string][]interface{}
}

func (b *fakePostBlog) PostID(url string) string {
	return "1"
}

func (b *fakePostBlog) Update(id string, replace, add, delete map[string][]interface{}) error {
	b.replaces[id] = append(b.replaces[id], replace)
	b.adds[id] = append(b.adds[id], add)
	b.deletes[id] = append(b.deletes[id], delete)

	return nil
}

func (b *fakePostBlog) SetNextPage(name string) error {
	return nil
}

func (b *fakePostBlog) Create(data map[string][]interface{}) (map[string][]interface{}, error) {
	b.datas = append(b.datas, data)

	return map[string][]interface{}{"url": {"http://example.com/blog/p/1"}}, nil
}

func (b *fakePostBlog) RenderPost(data map[string][]interface{}) error {
	return nil
}

func (b *fakePostBlog) Rerender(url string) error {
	return nil
}

func TestPostEntry(t *testing.T) {
	assert := assert.New(t)
	blog := &fakePostBlog{}

	s := httptest.NewServer(postHandler(blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"h":            {"entry"},
		"content":      {"This is a test"},
		"category[]":   {"test", "ignore"},
		"mp-something": {"what"},
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

		_, ok := data["mp-something"]
		assert.False(ok)
	}
}

func TestPostEntryJSON(t *testing.T) {
	assert := assert.New(t)
	blog := &fakePostBlog{}

	s := httptest.NewServer(postHandler(blog))
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
	assert.Equal("http://example.com/blog/p/1", resp.Header.Get("Location"))

	if assert.Len(blog.datas, 1) {
		data := blog.datas[0]

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
	blog := &fakePostBlog{
		adds:     map[string][]map[string][]interface{}{},
		deletes:  map[string][]map[string][]interface{}{},
		replaces: map[string][]map[string][]interface{}{},
	}

	s := httptest.NewServer(postHandler(blog))
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

	replace, ok := blog.replaces["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(replace, 1) {
		assert.Equal("hello moon", replace[0]["content"][0])
	}

	add, ok := blog.adds["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(add, 1) {
		assert.Equal("http://somewhere.com", add[0]["syndication"][0])
	}

	delete, ok := blog.deletes["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(delete, 1) {
		assert.Equal("this", delete[0]["not-important"][0])
	}
}
