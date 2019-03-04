package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

type goodHandler struct {
	OK bool
}

func (h *goodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.OK = true
}

type meHandler struct {
	Token string
	Me    string
}

func (h *meHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/verify" {
		if r.Header.Get("Authorization") == "Bearer "+h.Token {
			fmt.Fprint(w, `{
        "me": "`+h.Me+`",
        "client_id": "http://client.example.com/",
        "scope": "create"
      }`)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		fmt.Fprint(w, `
      <link rel="authorization_endpoint" href="/" />
      <link rel="token_endpoint" href="/verify" />
    `)
	}
}

func TestAuthenticate(t *testing.T) {
	assert := assert.New(t)
	good := &goodHandler{}
	me := &meHandler{Token: "abcde"}

	meServer := httptest.NewServer(me)
	defer meServer.Close()
	me.Me = meServer.URL

	s := httptest.NewServer(Authenticate(meServer.URL, "create", good))
	defer s.Close()

	_, err := http.Get(s.URL + "?access_token=abcde")
	assert.Nil(err)

	assert.True(good.OK)
}

func TestAuthenticateMissingScope(t *testing.T) {
	assert := assert.New(t)
	good := &goodHandler{}
	me := &meHandler{Token: "abcde"}

	meServer := httptest.NewServer(me)
	defer meServer.Close()
	me.Me = meServer.URL

	s := httptest.NewServer(Authenticate(meServer.URL, "edit", good))
	defer s.Close()

	_, err := http.Get(s.URL + "?access_token=abcde")
	assert.Nil(err)

	assert.False(good.OK)
}

func TestAuthenticateNotMe(t *testing.T) {
	assert := assert.New(t)
	good := &goodHandler{}
	me := &meHandler{Token: "abcde"}

	meServer := httptest.NewServer(me)
	defer meServer.Close()
	me.Me = "http://who.example.com"

	s := httptest.NewServer(Authenticate(meServer.URL, "edit", good))
	defer s.Close()

	_, err := http.Get(s.URL + "?access_token=abcde")
	assert.Nil(err)

	assert.False(good.OK)
}

func TestAuthenticatedBadToken(t *testing.T) {
	assert := assert.New(t)
	good := &goodHandler{}
	me := &meHandler{Token: "abcde"}

	meServer := httptest.NewServer(me)
	defer meServer.Close()
	me.Me = meServer.URL

	s := httptest.NewServer(Authenticate(meServer.URL, "create", good))
	defer s.Close()

	_, err := http.Get(s.URL + "?access_token=xyz")
	assert.Nil(err)

	assert.False(good.OK)
}
