package webmention

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
)

const waitTime = 5 * time.Millisecond

type fakeMicropubReader struct{}

func (b *fakeMicropubReader) Post(url string) (map[string][]interface{}, error) {
	if url != "http://example.com/weblog/post-id" {
		return map[string][]interface{}{}, errors.New("what is that")
	}

	return map[string][]interface{}{}, nil
}

type fakeNotifier struct {
	ch chan string
}

func (b *fakeNotifier) PostChanged(url string) error {
	b.ch <- url
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

func TestMention(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeNotifier{ch: make(chan string, 1)}
	mr := &fakeMicropubReader{}

	source := httptest.NewServer(stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`))
	defer source.Close()

	endpoint, reader, _ := Endpoint(db, mr, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case url := <-blog.ch:
		grouped, _ := reader.ForPost(url)

		if assert.Len(grouped, 1) {
			assert.Equal(map[string][]interface{}{
				"name":        {"A reply to some post"},
				"in-reply-to": {"http://example.com/weblog/post-id"},
				"hx-target":   {"http://example.com/weblog/post-id"},
			}, grouped[0].Properties)
		}
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWhenPostUpdated(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeNotifier{ch: make(chan string, 1)}
	mr := &fakeMicropubReader{}

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

	endpoint, reader, _ := Endpoint(db, mr, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case url := <-blog.ch:
		grouped, _ := reader.ForPost(url)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, grouped[0].Properties)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}

	resp, err = http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case url := <-blog.ch:
		grouped, _ := reader.ForPost(url)

		assert.Equal(map[string][]interface{}{
			"name":        {"A great reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, grouped[0].Properties)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWithHCardAndHEntry(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeNotifier{ch: make(chan string, 1)}
	mr := &fakeMicropubReader{}

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

	endpoint, reader, _ := Endpoint(db, mr, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case url := <-blog.ch:
		grouped, _ := reader.ForPost(url)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, grouped[0].Properties)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWithoutMicroformats(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeNotifier{ch: make(chan string, 1)}
	mr := &fakeMicropubReader{}

	source := httptest.NewServer(stringHandler(`
<p>
  Just a link to <a href="http://example.com/weblog/post-id">this post</a>.
</p>
`))
	defer source.Close()

	endpoint, reader, _ := Endpoint(db, mr, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case url := <-blog.ch:
		grouped, _ := reader.ForPost(url)

		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
		}, grouped[0].Properties)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionOfDeletedPost(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeNotifier{ch: make(chan string, 1)}
	mr := &fakeMicropubReader{}

	source := httptest.NewServer(goneHandler())
	defer source.Close()

	endpoint, reader, _ := Endpoint(db, mr, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case url := <-blog.ch:
		grouped, _ := reader.ForPost(url)

		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
			"gone":      {true},
		}, grouped[0].Properties)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}
