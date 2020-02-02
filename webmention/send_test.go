package webmention

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hawx.me/code/assert"
)

type req struct {
	source, target string
}

func TestSendLinkHeaderOtherHost(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)

	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs <- req{r.FormValue("source"), r.FormValue("target")}
	}))
	defer endpoint.Close()

	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", "<"+endpoint.URL+"/webmention>; rel=webmention")
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkHeaderAbsolute(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Set("Link", "<"+target.URL+"/webmention>; rel=webmention")
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkHeaderRelative(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Set("Link", "</webmention>; rel=webmention")
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkTagAbsolute(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<link rel="webmention" href="`+target.URL+`/webmention" />`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkTagRelative(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<link rel="webmention" href="/webmention" />`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendATagAbsolute(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<a rel="webmention" href="`+target.URL+`/webmention">webmention</a>`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendATagRelative(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<a rel="webmention" href="/webmention">webmention</a>`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkHeaderRelQuoted(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Set("Link", `</webmention>; rel="webmention"`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkTagMultipleValues(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<link rel="some webmention others" href="/webmention" />`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkHeaderRelMultipleValues(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Set("Link", `</webmention>; rel="some webmention and others"`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendMultipleMethods(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Set("Link", `</webmention>; rel="webmention"`)
		io.WriteString(w, `<html>
<head>
  <link rel="webmention" href="/bad" />
</head>
<body>
  <a href="/worse" rel="webmention">webmention</a>
</body>
</html>`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendToSelf(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<link rel="webmention" href="" />`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendToFirst(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<a href="/webmention" rel="webmention">
<link rel="webmention" href="/bad" />`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendMultipleLinkHeaders(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Add("Link", `</bad>; rel="not-webmention"`)
		w.Header().Add("Link", `</webmention>; rel="webmention"`)
		w.Header().Add("Link", `</worse>; rel="really-not-webmention"`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkHeaderWithMultipleLinks(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		w.Header().Set("Link", `</bad>; rel="not-webmention", </webmention>; rel="webmention", </worse>; rel="really-not-webmention"`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkWithNoHref(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `
<link rel="webmention" />
<a href="/webmention" rel="webmention">webmention</a>
`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendLinkWithQueryString(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" && r.URL.RawQuery == "query=yes" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
		}

		io.WriteString(w, `<link rel="webmention" href="/webmention?query=yes" />`)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestSendToRedirect(t *testing.T) {
	assert := assert.New(t)
	reqs := make(chan req, 1)
	var target *httptest.Server

	target = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/webmention" {
			reqs <- req{r.FormValue("source"), r.FormValue("target")}
			return
		}

		if r.URL.Path == "/redirect" {
			io.WriteString(w, `<link rel="webmention" href="/webmention" />`)
			return
		}

		http.Redirect(w, r, "/redirect", http.StatusFound)
	}))
	defer target.Close()

	err := Send("http://example.com/my-post", target.URL)
	assert.Nil(err)

	select {
	case r := <-reqs:
		assert.Equal("http://example.com/my-post", r.source)
		assert.Equal(target.URL, r.target)
	case <-time.After(10 * time.Millisecond):
		assert.Fail("timed out")
	}
}
