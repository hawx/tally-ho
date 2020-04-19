package webmention

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"hawx.me/code/assert"
)

const waitTime = 5 * time.Millisecond

type mention struct {
	source string
	data   map[string][]interface{}
}

type fakeBlog struct {
	ch chan mention
}

func (b *fakeBlog) BaseURL() string {
	return "http://example.com/"
}

func (b *fakeBlog) Entry(url string) (map[string][]interface{}, error) {
	if url != "http://example.com/weblog/post-id" {
		return map[string][]interface{}{}, errors.New("what is that")
	}

	return map[string][]interface{}{}, nil
}

func (b *fakeBlog) Mention(source string, data map[string][]interface{}) error {
	b.ch <- mention{source, data}
	return nil
}

func stringHandler(s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	}
}

func goneHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}
}

func sequenceHandlers(hs ...http.Handler) http.HandlerFunc {
	index := 0

	return func(w http.ResponseWriter, r *http.Request) {
		if index >= len(hs) {
			w.WriteHeader(999)
			return
		}

		hs[index].ServeHTTP(w, r)
		index++
	}
}

func newFormRequest(qs url.Values) *http.Request {
	req := httptest.NewRequest("POST", "http://localhost/", strings.NewReader(qs.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestMention(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWhenPostUpdated(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(sequenceHandlers(stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`), stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A great reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`)))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}

	req = newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp = w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A great reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWithHCardAndHEntry(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<div class="h-card">
  <p class="p-name">John Doe</p>
</div>

<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWithoutMicroformats(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<p>
  Just a link to <a href="http://example.com/weblog/post-id">this post</a>.
</p>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionOfDeletedPost(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(goneHandler())
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
			"hx-gone":   {true},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}
