package silos

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestGithub(t *testing.T) {
	rs := make(chan string, 1)

	s := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" && r.URL.Path == "/user" {
				w.Write([]byte(`{"login": "test_user"}`))
				return
			}

			if r.Method == "PUT" && r.URL.Path == "/user/starred/person/repo" {
				rs <- "starred"
				w.WriteHeader(http.StatusNoContent)
				return
			}
		},
	))
	defer s.Close()

	github, err := Github(GithubOptions{
		BaseURL: s.URL + "/",
	})
	if !assert.Nil(t, err) {
		return
	}

	assert.Equal(t, "@test_user on github", github.Name())

	t.Run("like", func(t *testing.T) {
		assert := assert.New(t)

		location, err := github.Create(map[string][]interface{}{
			"hx-kind": {"like"},
			"like-of": {"https://github.com/person/repo/"},
		})

		assert.Nil(err)
		assert.Equal("https://github.com/person/repo/", location)

		select {
		case r := <-rs:
			assert.Equal("starred", r)
		case <-time.After(time.Second):
			t.Fatal("expected like")
		}
	})
}
