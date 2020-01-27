package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"hawx.me/code/assert"
)

func testCases(queryAdd, headerAdd string) map[string]func(string) (*http.Response, error) {
	return map[string]func(string) (*http.Response, error){
		"query": func(u string) (*http.Response, error) {
			return http.Get(u + queryAdd)
		},
		"header": func(u string) (*http.Response, error) {
			req, _ := http.NewRequest("GET", u, nil)
			req.Header.Add("Authorization", "Bearer "+headerAdd)
			return http.DefaultClient.Do(req)
		},
	}
}

type goodHandler struct {
	OK bool
}

func (h *goodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.OK = true
}

type scopedHandler struct {
	OK bool
}

func (h *scopedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.OK = HasScope(w, r, "create")
}

type meHandler struct {
	Token string
	Me    string
	Scope string
}

func (h *meHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	scope := h.Scope
	if scope == "" {
		scope = "create"
	}

	if r.URL.Path == "/verify" {
		if r.Header.Get("Authorization") == "Bearer "+h.Token {
			fmt.Fprint(w, `{
        "me": "`+h.Me+`",
        "client_id": "http://client.example.com/",
        "scope": "`+scope+`"
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
	for name, f := range testCases("?access_token=abcde", "abcde") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &goodHandler{}
			me := &meHandler{Token: "abcde"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = meServer.URL

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			_, err := f(s.URL)
			assert.Nil(err)

			assert.True(good.OK)
		})
	}
}

func TestAuthenticateMissingScope(t *testing.T) {
	for name, f := range testCases("?access_token=abcde", "abcde") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &goodHandler{}
			me := &meHandler{Token: "abcde"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = meServer.URL

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			resp, err := f(s.URL)
			assert.Nil(err)
			assert.Equal(http.StatusOK, resp.StatusCode)

			assert.True(good.OK)
		})
	}
}

func TestAuthenticateNotMe(t *testing.T) {
	for name, f := range testCases("?access_token=abcde", "abcde") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &goodHandler{}
			me := &meHandler{Token: "abcde"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = "http://who.example.com"

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			resp, err := f(s.URL)
			assert.Nil(err)
			assert.Equal(http.StatusForbidden, resp.StatusCode)

			assert.False(good.OK)
		})
	}
}

func TestAuthenticatedBadToken(t *testing.T) {
	for name, f := range testCases("?access_token=xyz", "xyz") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &goodHandler{}
			me := &meHandler{Token: "abcde"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = meServer.URL

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			resp, err := f(s.URL)
			assert.Nil(err)
			assert.Equal(http.StatusForbidden, resp.StatusCode)

			assert.False(good.OK)
		})
	}
}

func TestAuthenticatedMissingToken(t *testing.T) {
	for name, f := range testCases("", "") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &goodHandler{}
			me := &meHandler{Token: "abcde"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = meServer.URL

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			resp, err := f(s.URL)
			assert.Nil(err)
			assert.Equal(http.StatusUnauthorized, resp.StatusCode)

			assert.False(good.OK)
		})
	}
}

func TestScopedAuthenticate(t *testing.T) {
	for name, f := range testCases("?access_token=abcde", "abcde") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &scopedHandler{}
			me := &meHandler{Token: "abcde"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = meServer.URL

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			resp, err := f(s.URL)
			assert.Nil(err)
			assert.Equal(http.StatusOK, resp.StatusCode)

			assert.True(good.OK)
		})
	}
}

func TestScopedAuthenticateMissingScope(t *testing.T) {
	for name, f := range testCases("?access_token=abcde", "abcde") {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)
			good := &scopedHandler{}
			me := &meHandler{Token: "abcde", Scope: "not-create"}

			meServer := httptest.NewServer(me)
			defer meServer.Close()
			me.Me = meServer.URL

			s := httptest.NewServer(Only(meServer.URL, good))
			defer s.Close()

			resp, err := f(s.URL)
			assert.Nil(err)
			assert.Equal(http.StatusUnauthorized, resp.StatusCode)

			assert.False(good.OK)
		})
	}
}
