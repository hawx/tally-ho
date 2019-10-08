package syndicate

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hawx.me/code/assert"
)

type Req struct {
	r    *http.Request
	body []byte
}

func TestTwitter(t *testing.T) {
	assert := assert.New(t)

	rs := make(chan Req, 1)

	s := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				body, err := ioutil.ReadAll(r.Body)
				assert.Nil(err)

				rs <- Req{r, body}

				w.Write([]byte(`{
  "id": 1050118621198921700,
  "id_str": "1050118621198921728",
  "user": {
    "url": "https://twitter.com/testing"
  }
}`))
			},
		),
	)
	defer s.Close()

	twitter := Twitter(TwitterOptions{
		BaseURL:           s.URL,
		ConsumerKey:       "consumer-key",
		ConsumerSecret:    "consumer-secret",
		AccessToken:       "access-token",
		AccessTokenSecret: "access-token-secret",
	})

	location, err := twitter.Create(map[string][]interface{}{
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
}
