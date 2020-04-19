package micropub

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

func withScope(scope string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "__hawx.me/code/tally-ho:Scopes__", []string{scope})))
	})
}

type fakePostDB struct {
	datas                   []map[string][]interface{}
	replaces, adds, deletes map[string][]map[string][]interface{}
	deleteAlls              map[string][][]string
	deleted                 []string
	undeleted               []string
}

func (b *fakePostDB) Create(data map[string][]interface{}) (string, error) {
	b.datas = append(b.datas, data)

	return "http://example.com/blog/p/1", nil
}

func (b *fakePostDB) Update(
	id string,
	replace, add, delete map[string][]interface{},
	deleteAlls []string,
) error {
	b.replaces[id] = append(b.replaces[id], replace)
	b.adds[id] = append(b.adds[id], add)
	b.deletes[id] = append(b.deletes[id], delete)
	b.deleteAlls[id] = append(b.deleteAlls[id], deleteAlls)

	return nil
}

func (b *fakePostDB) Delete(url string) error {
	b.deleted = append(b.deleted, url)
	return nil
}

func (b *fakePostDB) Undelete(url string) error {
	b.undeleted = append(b.undeleted, url)
	return nil
}

type fakeFileWriter struct {
	data []string
}

func (fw *fakeFileWriter) WriteFile(name, contentType string, r io.Reader) (string, error) {
	data, _ := ioutil.ReadAll(r)
	fw.data = append(fw.data, string(data))

	return "http://example.com/" + name, nil
}

func newFormRequest(qs url.Values) *http.Request {
	req := httptest.NewRequest("POST", "http://localhost/", strings.NewReader(qs.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func newJSONRequest(body string) *http.Request {
	req := httptest.NewRequest("POST", "http://localhost/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return req
}

type multipartFile struct {
	key, name, value string
}

func newMultipartRequest(fields url.Values, files []multipartFile) *http.Request {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, values := range fields {
		for _, value := range values {
			part, err := writer.CreateFormField(key)
			if err != nil {
				panic(err)
			}
			io.WriteString(part, value)
		}
	}
	for _, file := range files {
		part, err := writer.CreateFormFile(file.key, file.name)
		if err != nil {
			panic(err)
		}
		io.WriteString(part, file.value)
	}

	if err := writer.Close(); err != nil {
		panic(err)
	}

	req := httptest.NewRequest("POST", "http://localhost/", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestPostEntry(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"h":            {"entry"},
			"content":      {"This is a test"},
			"category[]":   {"test", "ignore"},
			"mp-something": {"what"},
			"url":          {"what"},
		}),
		"json": newJSONRequest(`{
  "type": ["h-entry"],
  "properties": {
    "content": ["This is a test"],
    "category": ["test", "ignore"],
    "mp-something": ["what"],
    "url": ["http://what"]
  }
}`),
		"multipart-form": newMultipartRequest(url.Values{
			"h":            {"entry"},
			"content":      {"This is a test"},
			"category[]":   {"test", "ignore"},
			"mp-something": {"what"},
			"url":          {"what"},
		}, nil),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			blog := &fakePostDB{}

			handler := withScope("create", postHandler(blog, nil))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			resp := w.Result()

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
		})
	}
}

func TestPostEntryMissingScope(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"h":            {"entry"},
			"content":      {"This is a test"},
			"category[]":   {"test", "ignore"},
			"mp-something": {"what"},
			"url":          {"what"},
		}),
		"json": newJSONRequest(`{
  "type": ["h-entry"],
  "properties": {
    "content": ["This is a test"],
    "category": ["test", "ignore"],
    "mp-something": ["what"],
    "url": ["http://what"]
  }
}`),
		"multipart-form": newMultipartRequest(url.Values{
			"h":            {"entry"},
			"content":      {"This is a test"},
			"category[]":   {"test", "ignore"},
			"mp-something": {"what"},
			"url":          {"what"},
		}, nil),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			blog := &fakePostDB{}

			handler := postHandler(blog, nil)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			resp := w.Result()

			assert.Equal(http.StatusUnauthorized, resp.StatusCode)
			assert.Len(blog.datas, 0)
		})
	}
}

func TestPostEntryMultipartFormWithMedia(t *testing.T) {
	for _, key := range []string{"photo", "video", "audio"} {
		t.Run(key, func(t *testing.T) {
			assert := assert.New(t)
			file := "this is an image"
			db := &fakePostDB{}
			fw := &fakeFileWriter{}

			handler := withScope("create", postHandler(db, fw))

			req := newMultipartRequest(url.Values{
				"h":       {"entry"},
				"content": {"This is a test"},
			}, []multipartFile{{key, "whatever.png", file}})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusCreated, resp.StatusCode)
			assert.Equal("http://example.com/blog/p/1", resp.Header.Get("Location"))

			if assert.Len(db.datas, 1) {
				data := db.datas[0]

				assert.Equal("entry", data["h"][0])
				assert.Equal("This is a test", data["content"][0])
				assert.Equal("http://example.com/whatever.png", data[key][0])
			}

			if assert.Len(fw.data, 1) {
				assert.Equal(file, fw.data[0])
			}
		})
	}
}

func TestPostEntryMultipartFormWithMediaMissingScope(t *testing.T) {
	for _, key := range []string{"photo", "video", "audio"} {
		t.Run(key, func(t *testing.T) {
			assert := assert.New(t)
			file := "this is an image"
			db := &fakePostDB{}
			fw := &fakeFileWriter{}

			handler := postHandler(db, fw)

			req := newMultipartRequest(url.Values{
				"h":       {"entry"},
				"content": {"This is a test"},
			}, []multipartFile{{key, "whatever.png", file}})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusUnauthorized, resp.StatusCode)
			assert.Len(db.datas, 0)
			assert.Len(fw.data, 0)
		})
	}
}

func TestPostEntryMultipartFormWithMultiplePhotos(t *testing.T) {
	for _, key := range []string{"photo", "video", "audio"} {
		t.Run(key, func(t *testing.T) {
			assert := assert.New(t)
			db := &fakePostDB{}
			fw := &fakeFileWriter{}

			handler := withScope("create", postHandler(db, fw))

			req := newMultipartRequest(url.Values{
				"h":       {"entry"},
				"content": {"This is a test"},
			}, []multipartFile{
				{key + "[]", "1.jpg", "the first file"},
				{key + "[]", "2.jpg", "the second image"},
			})

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusCreated, resp.StatusCode)
			assert.Equal("http://example.com/blog/p/1", resp.Header.Get("Location"))

			if assert.Len(db.datas, 1) {
				data := db.datas[0]

				assert.Equal("entry", data["h"][0])
				assert.Equal("This is a test", data["content"][0])
				assert.Equal("http://example.com/1.jpg", data[key][0])
				assert.Equal("http://example.com/2.jpg", data[key][1])
			}

			if assert.Len(fw.data, 2) {
				assert.Equal("the first file", fw.data[0])
				assert.Equal("the second image", fw.data[1])
			}
		})
	}
}

func TestPostEntryWithEmptyValues(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"h":          {"entry"},
			"content":    {"This is a test"},
			"category[]": {""},
		}),
		"json": newJSONRequest(`{
  "type": ["h-entry"],
  "properties": {
    "content": ["This is a test"],
    "category": []
  }
}`),
		"multipart-form": newMultipartRequest(url.Values{
			"h":          {"entry"},
			"content":    {"This is a test"},
			"category[]": {""},
		}, nil),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			blog := &fakePostDB{}

			handler := withScope("create", postHandler(blog, nil))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusCreated, resp.StatusCode)
			assert.Equal("http://example.com/blog/p/1", resp.Header.Get("Location"))

			if assert.Len(blog.datas, 1) {
				data := blog.datas[0]

				_, ok := data["category"]
				assert.False(ok)
			}
		})
	}
}

func TestUpdateEntry(t *testing.T) {
	assert := assert.New(t)
	db := &fakePostDB{
		adds:       map[string][]map[string][]interface{}{},
		deletes:    map[string][]map[string][]interface{}{},
		replaces:   map[string][]map[string][]interface{}{},
		deleteAlls: map[string][][]string{},
	}

	handler := withScope("update", postHandler(db, nil))

	req := newJSONRequest(`{
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
}`)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
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

func TestUpdateEntryMissingScope(t *testing.T) {
	assert := assert.New(t)
	db := &fakePostDB{
		adds:       map[string][]map[string][]interface{}{},
		deletes:    map[string][]map[string][]interface{}{},
		replaces:   map[string][]map[string][]interface{}{},
		deleteAlls: map[string][][]string{},
	}

	handler := postHandler(db, nil)

	req := newJSONRequest(`{
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
}`)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusUnauthorized, resp.StatusCode)

	_, ok := db.replaces["https://example.com/blog/p/100"]
	assert.False(ok)

	_, ok = db.adds["https://example.com/blog/p/100"]
	assert.False(ok)

	_, ok = db.deletes["https://example.com/blog/p/100"]
	assert.False(ok)
}

func TestUpdateEntryDelete(t *testing.T) {
	assert := assert.New(t)
	db := &fakePostDB{
		adds:       map[string][]map[string][]interface{}{},
		deletes:    map[string][]map[string][]interface{}{},
		replaces:   map[string][]map[string][]interface{}{},
		deleteAlls: map[string][][]string{},
	}

	handler := withScope("update", postHandler(db, nil))

	req := newJSONRequest(`{
  "action": "update",
  "url": "https://example.com/blog/p/100",
  "replace": {
    "content": ["hello moon"]
  },
  "add": {
    "syndication": ["http://somewhere.com"]
  },
  "delete": ["not-important"]
}`)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusNoContent, resp.StatusCode)

	replace, ok := db.replaces["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(replace, 1) {
		assert.Equal("hello moon", replace[0]["content"][0])
	}

	add, ok := db.adds["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(add, 1) {
		assert.Equal("http://somewhere.com", add[0]["syndication"][0])
	}

	delete, ok := db.deleteAlls["https://example.com/blog/p/100"]
	if assert.True(ok) && assert.Len(delete, 1) {
		assert.Equal("not-important", delete[0][0])
	}
}

func TestUpdateEntryInvalidDelete(t *testing.T) {
	handler := withScope("update", postHandler(nil, nil))

	testCases := map[string]string{
		"array with non-string": `[1]`,
		"map with non-array":    `{"this-key": "and-value"}`,
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			req := newJSONRequest(`{
  "action": "update",
  "url": "https://example.com/blog/p/100",
  "delete": ` + tc + `
}`)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestDeleteEntry(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"action": {"delete"},
			"url":    {"https://example.com/blog/p/1"},
		}),
		"json": newJSONRequest(`{"action": "delete", "url": "https://example.com/blog/p/1"}`),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			db := &fakePostDB{}

			handler := withScope("delete", postHandler(db, nil))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusNoContent, resp.StatusCode)

			if assert.Len(db.deleted, 1) {
				assert.Equal("https://example.com/blog/p/1", db.deleted[0])
			}
		})
	}
}

func TestDeleteEntryMissingScope(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"action": {"delete"},
			"url":    {"https://example.com/blog/p/1"},
		}),
		"json": newJSONRequest(`{"action": "delete", "url": "https://example.com/blog/p/1"}`),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			db := &fakePostDB{}

			handler := postHandler(db, nil)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusUnauthorized, resp.StatusCode)
			assert.Len(db.deleted, 0)
		})
	}
}

func TestUndeleteEntry(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"action": {"undelete"},
			"url":    {"https://example.com/blog/p/1"},
		}),
		"json": newJSONRequest(`{"action": "undelete", "url": "https://example.com/blog/p/1"}`),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			db := &fakePostDB{}

			handler := withScope("delete", postHandler(db, nil))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusNoContent, resp.StatusCode)

			if assert.Len(db.undeleted, 1) {
				assert.Equal("https://example.com/blog/p/1", db.undeleted[0])
			}
		})
	}
}

func TestUndeleteEntryMissingScope(t *testing.T) {
	testCases := map[string]*http.Request{
		"url-encoded-form": newFormRequest(url.Values{
			"action": {"undelete"},
			"url":    {"https://example.com/blog/p/1"},
		}),
		"json": newJSONRequest(`{"action": "undelete", "url": "https://example.com/blog/p/1"}`),
	}

	for name, req := range testCases {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			db := &fakePostDB{}

			handler := postHandler(db, nil)

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			resp := w.Result()
			assert.Equal(http.StatusUnauthorized, resp.StatusCode)
			assert.Len(db.undeleted, 0)
		})
	}
}
