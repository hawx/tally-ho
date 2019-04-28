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
	"hawx.me/code/numbersix"
)

func triplesToMap(triples []numbersix.Triple) (map[string][]interface{}, error) {
	properties := map[string][]interface{}{}
	for _, triple := range triples {
		var value interface{}
		if err := triple.Value(&value); err != nil {
			return properties, err
		}

		properties[triple.Predicate] = append(properties[triple.Predicate], value)
	}

	return properties, nil
}

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

	endpoint, _ := Endpoint(db, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	ndb, _ := numbersix.For(db, "mentions")

	time.Sleep(time.Millisecond)
	triples, _ := ndb.List(numbersix.All())
	data, _ := triplesToMap(triples)

	assert.Equal(map[string][]interface{}{
		"name":        {"A reply to some post"},
		"in-reply-to": {"http://example.com/weblog/post-id"},
		"hx-target":   {"http://example.com/weblog/post-id"},
	}, data)
}

func TestMentionWhenPostUpdated(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
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

	endpoint, _ := Endpoint(db, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	ndb, _ := numbersix.For(db, "mentions")

	time.Sleep(time.Millisecond)
	triples, _ := ndb.List(numbersix.All())
	data, _ := triplesToMap(triples)

	assert.Equal(map[string][]interface{}{
		"name":        {"A reply to some post"},
		"in-reply-to": {"http://example.com/weblog/post-id"},
		"hx-target":   {"http://example.com/weblog/post-id"},
	}, data)

	resp, err = http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	time.Sleep(time.Millisecond)
	triples, _ = ndb.List(numbersix.All())
	data, _ = triplesToMap(triples)

	assert.Equal(map[string][]interface{}{
		"name":        {"A great reply to some post"},
		"in-reply-to": {"http://example.com/weblog/post-id"},
		"hx-target":   {"http://example.com/weblog/post-id"},
	}, data)
}

func TestMentionWithHCardAndHEntry(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
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

	endpoint, _ := Endpoint(db, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	ndb, _ := numbersix.For(db, "mentions")

	time.Sleep(time.Millisecond)
	triples, _ := ndb.List(numbersix.All())
	data, _ := triplesToMap(triples)

	assert.Equal(map[string][]interface{}{
		"name":        {"A reply to some post"},
		"in-reply-to": {"http://example.com/weblog/post-id"},
		"hx-target":   {"http://example.com/weblog/post-id"},
	}, data)
}

func TestMentionWithoutMicroformats(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeMentionBlog{}

	source := httptest.NewServer(stringHandler(`
<p>
  Just a link to <a href="http://example.com/weblog/post-id">this post</a>.
</p>
`))
	defer source.Close()

	endpoint, _ := Endpoint(db, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	ndb, _ := numbersix.For(db, "mentions")

	time.Sleep(time.Millisecond)
	triples, _ := ndb.List(numbersix.All())
	data, _ := triplesToMap(triples)

	assert.Equal(map[string][]interface{}{
		"hx-target": {"http://example.com/weblog/post-id"},
	}, data)
}

func TestMentionOfDeletedPost(t *testing.T) {
	assert := assert.New(t)

	db, _ := sql.Open("sqlite3", "file::memory:")
	blog := &fakeMentionBlog{}

	source := httptest.NewServer(goneHandler())
	defer source.Close()

	endpoint, _ := Endpoint(db, blog)
	s := httptest.NewServer(endpoint)
	defer s.Close()

	resp, err := http.PostForm(s.URL, url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	ndb, _ := numbersix.For(db, "mentions")

	time.Sleep(time.Millisecond)
	triples, _ := ndb.List(numbersix.All())
	data, _ := triplesToMap(triples)

	assert.Equal(map[string][]interface{}{
		"hx-target": {"http://example.com/weblog/post-id"},
		"gone":      {true},
	}, data)
}
