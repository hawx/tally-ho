package websub

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"hawx.me/code/assert"
)

type fakeSubIter struct {
	current int
	subs    []fakeSub
}

func (i *fakeSubIter) Close() error { return nil }
func (i *fakeSubIter) Data() (string, string) {
	here := i.subs[i.current-1]

	return here.callback, here.secret
}
func (i *fakeSubIter) Err() error { return nil }
func (i *fakeSubIter) Next() bool {
	if i.current < len(i.subs) {
		i.current += 1
		return true
	}

	return false
}

type fakeSub struct {
	callback  string
	topic     string
	expiresAt time.Time
	secret    string
}

type fakeHubStore struct {
	subs   []fakeSub
	unsubs []fakeSub
}

func (s *fakeHubStore) Subscribe(callback, topic string, expiresAt time.Time, secret string) error {
	s.subs = append(s.subs, fakeSub{callback, topic, expiresAt, secret})
	return nil
}

func (s *fakeHubStore) Subscribers(topic string) SubscribersIter {
	return &fakeSubIter{current: 0, subs: s.subs}
}

func (s *fakeHubStore) Unsubscribe(callback, topic string) error {
	s.unsubs = append(s.unsubs, fakeSub{callback, topic, time.Now(), ""})
	return nil
}

func TestSubscribe(t *testing.T) {
	assert := assert.New(t)
	challenge := []byte{1, 2, 3, 4}

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)
	hub.generator = func() ([]byte, error) {
		return challenge, nil
	}

	h := httptest.NewServer(hub)
	defer h.Close()

	verification := make(chan url.Values, 1)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/unguessable-path-unique-per-subscription" {
			verification <- r.URL.Query()
			w.Write(challenge)
		}
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {s.URL + "/unguessable-path-unique-per-subscription?keep=me"},
		"hub.mode":     {"subscribe"},
		"hub.topic":    {"http://example.com/category/cats"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-verification:
		assert.Equal("me", v.Get("keep"))
		assert.Equal("subscribe", v.Get("hub.mode"))
		assert.Equal("http://example.com/category/cats", v.Get("hub.topic"))
		assert.Equal(string(challenge), v.Get("hub.challenge"))
		assert.Equal("864000", v.Get("hub.lease_seconds"))
	case <-time.After(time.Millisecond):
		assert.Fail("timed out")
	}

	if assert.Len(store.subs, 1) {
		sub := store.subs[0]
		assert.Equal(s.URL+"/unguessable-path-unique-per-subscription?keep=me", sub.callback)
		assert.Equal("http://example.com/category/cats", sub.topic)
		assert.WithinDuration(time.Now().Add(864000*time.Second), sub.expiresAt, time.Second)
		assert.Equal("", sub.secret)
	}
}

func TestSubscribeWithSecret(t *testing.T) {
	assert := assert.New(t)
	challenge := []byte{1, 2, 3, 4}

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)
	hub.generator = func() ([]byte, error) {
		return challenge, nil
	}

	h := httptest.NewServer(hub)
	defer h.Close()

	verification := make(chan url.Values, 1)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/unguessable-path-unique-per-subscription" {
			verification <- r.URL.Query()
			w.Write(challenge)
		}
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {s.URL + "/unguessable-path-unique-per-subscription?keep=me"},
		"hub.mode":     {"subscribe"},
		"hub.topic":    {"http://example.com/category/cats"},
		"hub.secret":   {"catgifs"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-verification:
		assert.Equal("me", v.Get("keep"))
		assert.Equal("subscribe", v.Get("hub.mode"))
		assert.Equal("http://example.com/category/cats", v.Get("hub.topic"))
		assert.Equal(string(challenge), v.Get("hub.challenge"))
		assert.Equal("864000", v.Get("hub.lease_seconds"))
	case <-time.After(time.Millisecond):
		assert.Fail("timed out")
	}

	if assert.Len(store.subs, 1) {
		sub := store.subs[0]
		assert.Equal(s.URL+"/unguessable-path-unique-per-subscription?keep=me", sub.callback)
		assert.Equal("http://example.com/category/cats", sub.topic)
		assert.WithinDuration(time.Now().Add(864000*time.Second), sub.expiresAt, time.Second)
		assert.Equal("catgifs", sub.secret)
	}
}

func TestSubscribeWithLongSecret(t *testing.T) {
	assert := assert.New(t)
	challenge := []byte{1, 2, 3, 4}

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)
	hub.generator = func() ([]byte, error) {
		return challenge, nil
	}

	h := httptest.NewServer(hub)
	defer h.Close()

	verification := make(chan url.Values, 1)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/unguessable-path-unique-per-subscription" {
			verification <- r.URL.Query()
			w.Write(challenge)
		}
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {s.URL + "/unguessable-path-unique-per-subscription?keep=me"},
		"hub.mode":     {"subscribe"},
		"hub.topic":    {"http://example.com/category/cats"},
		"hub.secret":   {"12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890X"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestSubscribeWithSpecificLease(t *testing.T) {
	assert := assert.New(t)
	challenge := []byte{1, 2, 3, 4}

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)
	hub.generator = func() ([]byte, error) {
		return challenge, nil
	}

	h := httptest.NewServer(hub)
	defer h.Close()

	verification := make(chan url.Values, 1)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/unguessable-path-unique-per-subscription" {
			verification <- r.URL.Query()
			w.Write(challenge)
		}
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback":      {s.URL + "/unguessable-path-unique-per-subscription?keep=me"},
		"hub.mode":          {"subscribe"},
		"hub.topic":         {"http://example.com/category/cats"},
		"hub.lease_seconds": {strconv.Itoa(int((24 * time.Hour).Seconds()))},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-verification:
		assert.Equal("me", v.Get("keep"))
		assert.Equal("subscribe", v.Get("hub.mode"))
		assert.Equal("http://example.com/category/cats", v.Get("hub.topic"))
		assert.Equal(string(challenge), v.Get("hub.challenge"))
		assert.Equal("86400", v.Get("hub.lease_seconds"))
	case <-time.After(time.Millisecond):
		assert.Fail("timed out")
	}

	if assert.Len(store.subs, 1) {
		sub := store.subs[0]
		assert.Equal(s.URL+"/unguessable-path-unique-per-subscription?keep=me", sub.callback)
		assert.Equal("http://example.com/category/cats", sub.topic)
		assert.WithinDuration(time.Now().Add(86400*time.Second), sub.expiresAt, time.Second)
	}
}

func TestSubscribeWithTooLongLease(t *testing.T) {
	assert := assert.New(t)
	challenge := []byte{1, 2, 3, 4}

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)
	hub.generator = func() ([]byte, error) {
		return challenge, nil
	}

	h := httptest.NewServer(hub)
	defer h.Close()

	verification := make(chan url.Values, 1)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/unguessable-path-unique-per-subscription" {
			verification <- r.URL.Query()
			w.Write(challenge)
		}
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback":      {s.URL + "/unguessable-path-unique-per-subscription?keep=me"},
		"hub.mode":          {"subscribe"},
		"hub.topic":         {"http://example.com/category/cats"},
		"hub.lease_seconds": {strconv.Itoa(int((29 * 24 * time.Hour).Seconds()))},
	})
	assert.Nil(err)
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case v := <-verification:
		assert.Equal("me", v.Get("keep"))
		assert.Equal("subscribe", v.Get("hub.mode"))
		assert.Equal("http://example.com/category/cats", v.Get("hub.topic"))
		assert.Equal(string(challenge), v.Get("hub.challenge"))
		assert.Equal("2419200", v.Get("hub.lease_seconds"))
	case <-time.After(time.Millisecond):
		assert.Fail("timed out")
	}

	if assert.Len(store.subs, 1) {
		sub := store.subs[0]
		assert.Equal(s.URL+"/unguessable-path-unique-per-subscription?keep=me", sub.callback)
		assert.Equal("http://example.com/category/cats", sub.topic)
		assert.WithinDuration(time.Now().Add(2419200*time.Second), sub.expiresAt, time.Second)
	}
}

func TestSubscribeWhenRespondingWithWrongChallenge(t *testing.T) {
	assert := assert.New(t)

	h := httptest.NewServer(New("", nil))
	defer h.Close()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/unguessable-path-unique-per-subscription" {
			w.Write([]byte("this-is-not-the-challenge"))
		}
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {s.URL + "/unguessable-path-unique-per-subscription"},
		"hub.mode":     {"subscribe"},
		"hub.topic":    {"http://example.com/category/cats"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestSubscribeNotPostRequest(t *testing.T) {
	assert := assert.New(t)

	h := httptest.NewServer(New("", nil))
	defer h.Close()

	resp, err := http.Get(h.URL)
	assert.Nil(err)
	assert.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestSubscribeBadCallback(t *testing.T) {
	assert := assert.New(t)

	h := httptest.NewServer(New("", nil))
	defer h.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {"this-aint-a-url"},
		"hub.mode":     {"subscribe"},
		"hub.topic":    {"http://example.com/category/cats"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestSubscribeBadMode(t *testing.T) {
	assert := assert.New(t)

	h := httptest.NewServer(New("", nil))
	defer h.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {"http://example.com/callback"},
		"hub.mode":     {"what"},
		"hub.topic":    {"http://example.com/category/cats"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestSubscribeBadVerificationResponse(t *testing.T) {
	assert := assert.New(t)

	h := httptest.NewServer(New("", nil))
	defer h.Close()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer s.Close()

	resp, err := http.PostForm(h.URL, url.Values{
		"hub.callback": {s.URL},
		"hub.mode":     {"subscribe"},
		"hub.topic":    {"http://example.com/category/cats"},
	})
	assert.Nil(err)
	assert.Equal(http.StatusBadRequest, resp.StatusCode)
}

func TestPublish(t *testing.T) {
	assert := assert.New(t)

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)

	type request struct {
		body    string
		headers http.Header
	}

	req := make(chan request, 1)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		req <- request{string(data), r.Header}
	}))
	defer s.Close()

	c := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plainest")
		w.Write([]byte("i-am-content"))
	}))
	defer c.Close()

	store.Subscribe(s.URL, c.URL, time.Now().Add(time.Second), "")

	err := hub.Publish(c.URL)
	assert.Nil(err)

	select {
	case r := <-req:
		assert.Equal("i-am-content", r.body)
		assert.Equal("text/plainest", r.headers.Get("Content-Type"))
		assert.Equal(`<http://hub.example.com/>; rel="hub", <`+c.URL+`>; rel="self"`, r.headers.Get("Link"))
	case <-time.After(time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestPublishWithSignature(t *testing.T) {
	assert := assert.New(t)

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)

	type request struct {
		body    string
		headers http.Header
	}

	req := make(chan request, 1)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := ioutil.ReadAll(r.Body)
		req <- request{string(data), r.Header}
	}))
	defer s.Close()

	c := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plainest")
		w.Write([]byte("i-am-content"))
	}))
	defer c.Close()

	store.Subscribe(s.URL, c.URL, time.Now().Add(time.Second), "catgifs")

	err := hub.Publish(c.URL)
	assert.Nil(err)

	select {
	case r := <-req:
		assert.Equal("i-am-content", r.body)
		assert.Equal("text/plainest", r.headers.Get("Content-Type"))
		assert.Equal("sha512=1c02e2bb4ac82bee90b65021299a87c2c8691f2b4fd5e72f2f8c169ee6732ec6baa4276bfe22fd9da1f4ea07e05c64229878a45256884fd507234dfafd3f6c81", r.headers.Get("X-Hub-Signature"))
		assert.Equal(`<http://hub.example.com/>; rel="hub", <`+c.URL+`>; rel="self"`, r.headers.Get("Link"))
	case <-time.After(time.Millisecond):
		assert.Fail("timed out")
	}
}

func TestPublishReturnsRedirect(t *testing.T) {
	assert := assert.New(t)

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)

	type request struct {
		body    string
		headers http.Header
	}

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail("should not be called")
	}))
	defer s2.Close()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s2.URL, http.StatusFound)
	}))
	defer s.Close()

	c := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plainest")
		w.Write([]byte("i-am-content"))
	}))
	defer c.Close()

	store.Subscribe(s.URL, c.URL, time.Now().Add(time.Second), "")

	err := hub.Publish(c.URL)
	assert.Nil(err)
}

func TestPublishReturnsGone(t *testing.T) {
	assert := assert.New(t)

	store := &fakeHubStore{}
	hub := New("http://hub.example.com/", store)

	type request struct {
		body    string
		headers http.Header
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}))
	defer s.Close()

	c := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plainest")
		w.Write([]byte("i-am-content"))
	}))
	defer c.Close()

	store.Subscribe(s.URL, c.URL, time.Now().Add(time.Second), "")

	err := hub.Publish(c.URL)
	assert.Nil(err)

	if assert.Len(store.unsubs, 1) {
		unsub := store.unsubs[0]

		assert.Equal(s.URL, unsub.callback)
		assert.Equal(c.URL, unsub.topic)
	}
}
