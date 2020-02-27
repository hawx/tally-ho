package silos

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestFlickr(t *testing.T) {
	qs := make(chan url.Values, 1)

	s := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" && r.FormValue("method") == "flickr.test.login" {
					w.Write([]byte(`{"user":{"username":{"_content":"someone"}}}`))
					return
				}

				if r.Method != "POST" {
					return
				}

				r.ParseForm()
				qs <- r.PostForm

				if r.FormValue("method") == "flickr.photos.comments.addComment" {
					w.Write([]byte(`{"comment":{"permalink":"https://www.flickr.com/someone/photos/43333233#comment-123"}}`))
				}
			},
		),
	)
	defer s.Close()

	flickr, err := Flickr(FlickrOptions{
		BaseURL: s.URL,
	})
	if !assert.Nil(t, err) {
		return
	}

	t.Run("reply", func(t *testing.T) {
		assert := assert.New(t)

		location, err := flickr.Create(map[string][]interface{}{
			"hx-kind":     {"reply"},
			"in-reply-to": {"https://www.flickr.com/photos/someone/43324322/"},
			"content":     {"cool pic"},
		})

		assert.Nil(err)
		assert.Equal("https://www.flickr.com/someone/photos/43333233#comment-123", location)

		select {
		case q := <-qs:
			assert.Equal("flickr.photos.comments.addComment", q.Get("method"))
			assert.Equal("43324322", q.Get("photo_id"))
			assert.Equal("cool pic", q.Get("comment_text"))
		case <-time.After(time.Second):
			t.Fatal("expected request to be made within 1s")
		}
	})
}
