package silos

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
)

type Req struct {
	r    *http.Request
	body []byte
}

func TestTwitterCreate(t *testing.T) {
	rs := make(chan Req, 1)

	s := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/account/verify_credentials.json" {
					w.Write([]byte(`{"screen_name": "TwitterDev"}`))
					return
				}

				body, err := ioutil.ReadAll(r.Body)
				assert.Nil(t, err)

				rs <- Req{r, body}

				w.Write([]byte(`{
  "id": 1050118621198921700,
  "id_str": "1050118621198921728",
  "user": {
    "screen_name": "testing"
  }
}`))
			},
		),
	)
	defer s.Close()

	twitter, err := Twitter(TwitterOptions{
		BaseURL:           s.URL,
		ConsumerKey:       "consumer-key",
		ConsumerSecret:    "consumer-secret",
		AccessToken:       "access-token",
		AccessTokenSecret: "access-token-secret",
	})
	if !assert.Nil(t, err) {
		return
	}

	t.Run("like-string", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind": {"like"},
			"like-of": {"https://twitter.com/SomePerson/status/1234"},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/SomePerson/status/1234", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/favorites/create.json", r.URL.Path)
			assert.Equal("id=1234", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})

	t.Run("like-cite", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind": {"like"},
			"like-of": {map[string]interface{}{
				"type": []string{"h-cite"},
				"properties": map[string][]interface{}{
					"url": {"https://twitter.com/SomePerson/status/1234"},
				},
			}},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/SomePerson/status/1234", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/favorites/create.json", r.URL.Path)
			assert.Equal("id=1234", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})

	t.Run("like-cite-query", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind": {"like"},
			"like-of": {map[string]interface{}{
				"type": []string{"h-cite"},
				"properties": map[string][]interface{}{
					"url": {"https://twitter.com/SomePerson/status/1234?s=09"},
				},
			}},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/SomePerson/status/1234?s=09", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/favorites/create.json", r.URL.Path)
			assert.Equal("id=1234", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})

	t.Run("note-string", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind": {"note"},
			"content": {"This is my tweet"},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/testing/status/1050118621198921728", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/statuses/update.json", r.URL.Path)
			assert.Equal("status=This+is+my+tweet", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})

	t.Run("note-html-text", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind": {"note"},
			"content": {map[string]interface{}{
				"text": "This is my tweet",
				"html": "This is my html",
			}},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/testing/status/1050118621198921728", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/statuses/update.json", r.URL.Path)
			assert.Equal("status=This+is+my+tweet", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})

	t.Run("reply-string", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind":     {"reply"},
			"in-reply-to": {"https://twitter.com/SomePerson/status/1234"},
			"content":     {"This is my tweet"},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/testing/status/1050118621198921728", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/statuses/update.json", r.URL.Path)
			assert.Equal("in_reply_to_status_id=1234&status=%40SomePerson+This+is+my+tweet", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})

	t.Run("reply-cite", func(t *testing.T) {
		assert := assert.New(t)

		location, err := twitter.Create(map[string][]interface{}{
			"hx-kind": {"reply"},
			"in-reply-to": {map[string]interface{}{
				"type": []string{"h-cite"},
				"properties": map[string][]interface{}{
					"url": {"https://twitter.com/SomePerson/status/1234"},
				},
			}},
			"content": {map[string]interface{}{
				"text": "This is my tweet",
				"html": "wot",
			}},
		})

		assert.Nil(err)
		assert.Equal("https://twitter.com/testing/status/1050118621198921728", location)

		select {
		case req := <-rs:
			r, body := req.r, req.body

			assert.Equal("POST", r.Method)
			assert.Equal("/statuses/update.json", r.URL.Path)
			assert.Equal("in_reply_to_status_id=1234&status=%40SomePerson+This+is+my+tweet", string(body))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})
}

func TestTwitterResolveCite(t *testing.T) {
	qs := make(chan url.Values, 1)

	s := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/account/verify_credentials.json" {
					w.Write([]byte(`{"screen_name": "TwitterDev"}`))
					return
				}

				if r.URL.Path == "/statuses/show.json" {
					qs <- r.URL.Query()

					w.Write([]byte(`{
  "id": 1050118621198921700,
  "id_str": "1050118621198921728",
  "text": "Hey there",
  "user": {
    "url": "https://t.co/something",
    "name": "Test Thing",
    "screen_name": "testing"
  }
}`))
				}
			},
		),
	)
	defer s.Close()

	twitter, err := Twitter(TwitterOptions{
		BaseURL:           s.URL,
		ConsumerKey:       "consumer-key",
		ConsumerSecret:    "consumer-secret",
		AccessToken:       "access-token",
		AccessTokenSecret: "access-token-secret",
	})
	if !assert.Nil(t, err) {
		return
	}

	cite, err := twitter.ResolveCite("https://twitter.com/johndoe/status/1432")
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"name":    {"@testing's tweet"},
			"content": {"Hey there"},
			"url":     {"https://twitter.com/johndoe/status/1432"},
			"author": {
				map[string]interface{}{
					"type": []interface{}{"h-card"},
					"properties": map[string][]interface{}{
						"name":     {"Test Thing"},
						"url":      {"https://twitter.com/testing"},
						"nickname": {"@testing"},
					},
				},
			},
		},
	}, cite)

	select {
	case q := <-qs:
		assert.Equal(t, url.Values{
			"id":         {"1432"},
			"tweet_mode": {"extended"},
		}, q)
	case <-time.After(time.Millisecond):
		assert.Fail(t, "timed out")
	}
}
