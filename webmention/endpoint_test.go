package webmention

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
)

type fakeMention struct {
	source     string
	properties map[string][]interface{}
}

type fakeMentionBlog struct {
}

func (b *fakeMentionBlog) PostByURL(url string) (map[string][]interface{}, error) {
	if url != "http://example.com/weblog/post-id" {
		return map[string][]interface{}{}, errors.New("what is that")
	}

	return map[string][]interface{}{}, nil
}

func (b *fakeMentionBlog) MentionSourceAllowed(url string) bool {
	return true
}

type fakeWebmentionDB struct {
	ch chan fakeMention
}

func (db *fakeWebmentionDB) Upsert(source string, data map[string][]interface{}) error {
	db.ch <- fakeMention{source, data}
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

	db := &fakeWebmentionDB{ch: make(chan fakeMention, 1)}
	blog := &fakeMentionBlog{}

	source := httptest.NewServer(stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`))
	defer source.Close()

	s := httptest.NewServer(postHandler(db, blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-db.ch:
		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, v.properties)
	case <-time.After(time.Second):
		t.Fatal("failed to send mention")
	}
}

func TestMentionWhenPostUpdated(t *testing.T) {
	assert := assert.New(t)

	db := &fakeWebmentionDB{ch: make(chan fakeMention, 1)}
	blog := &fakeMentionBlog{}

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

	s := httptest.NewServer(postHandler(db, blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-db.ch:
		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, v.properties)
	case <-time.After(time.Second):
		t.Fatal("failed to send mention")
	}

	resp, err = http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-db.ch:
		assert.Equal(map[string][]interface{}{
			"name":        {"A great reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, v.properties)
	case <-time.After(time.Second):
		t.Fatal("failed to send mention")
	}
}

func TestMentionWithHCardAndHEntry(t *testing.T) {
	assert := assert.New(t)

	db := &fakeWebmentionDB{ch: make(chan fakeMention, 1)}
	blog := &fakeMentionBlog{}

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

	s := httptest.NewServer(postHandler(db, blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-db.ch:
		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, v.properties)
	case <-time.After(time.Second):
		t.Fatal("failed to send mention")
	}
}

func TestMentionWithoutMicroformats(t *testing.T) {
	assert := assert.New(t)

	db := &fakeWebmentionDB{ch: make(chan fakeMention, 1)}
	blog := &fakeMentionBlog{}

	source := httptest.NewServer(stringHandler(`
<p>
  Just a link to <a href="http://example.com/weblog/post-id">this post</a>.
</p>
`))
	defer source.Close()

	s := httptest.NewServer(postHandler(db, blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-db.ch:
		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
		}, v.properties)
	case <-time.After(time.Second):
		t.Fatal("failed to send mention")
	}
}

func TestMentionOfDeletedPost(t *testing.T) {
	assert := assert.New(t)

	db := &fakeWebmentionDB{ch: make(chan fakeMention, 1)}
	blog := &fakeMentionBlog{}

	source := httptest.NewServer(goneHandler())
	defer source.Close()

	s := httptest.NewServer(postHandler(db, blog))
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-db.ch:
		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
			"gone":      {true},
		}, v.properties)
	case <-time.After(time.Second):
		t.Fatal("failed to send mention")
	}
}
