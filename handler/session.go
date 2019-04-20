package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"io"
	"net/http"

	"github.com/gorilla/sessions"
	"hawx.me/code/indieauth"
)

type userSession struct {
	Me          string
	AccessToken string
	Micropub    string
	Media       string
}

func init() {
	gob.Register(userSession{})
}

type scopedSessions struct {
	me               string
	store            sessions.Store
	auth             *indieauth.AuthorizationConfig
	ends             indieauth.Endpoints
	defaultSignedOut http.Handler
	Root             string
}

func NewScopedSessions(me, secret string, auth *indieauth.AuthorizationConfig) (*scopedSessions, error) {
	if me == "" {
		return nil, errors.New("me must be non-empty")
	}

	byteSecret, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, err
	}

	endpoints, err := indieauth.FindEndpoints(me)
	if err != nil {
		return nil, err
	}

	return &scopedSessions{
		me:               me,
		store:            sessions.NewCookieStore(byteSecret),
		auth:             auth,
		ends:             endpoints,
		defaultSignedOut: http.NotFoundHandler(),
		Root:             "/",
	}, nil
}

func (s *scopedSessions) get(r *http.Request) (userSession, bool) {
	session, _ := s.store.Get(r, "session")

	user, ok := session.Values["user"].(userSession)
	return user, ok
}

func (s *scopedSessions) set(w http.ResponseWriter, r *http.Request, token indieauth.Token) {
	session, _ := s.store.Get(r, "session")

	// meURL, _ := url.Parse(token.Me)

	// resp, _ := http.Get(token.Me)
	// defer resp.Body.Close()
	// data := microformats.Parse(resp.Body, meURL)
	// micropubEndpoint := data.Rels["micropub"][0]

	// resp, _ = http.Get(micropubEndpoint + "?q=config") // TODO: parse the URL as could have query
	// defer resp.Body.Close()
	// var data2 map[string]string
	// json.NewDecoder(resp.Body).Decode(&data2)
	// mediaEndpoint := data2["media-endpoint"]
	session.Values["user"] = userSession{
		Me:          token.Me,
		AccessToken: token.AccessToken,
		Micropub:    "/micropub",
		Media:       "/media",
	}
	session.Save(r, w)
}

func (s *scopedSessions) setState(w http.ResponseWriter, r *http.Request) (string, error) {
	bytes := make([]byte, 32)

	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(bytes)

	session, _ := s.store.Get(r, "session")
	session.Values["state"] = state
	return state, session.Save(r, w)
}

func (s *scopedSessions) getState(r *http.Request) string {
	session, _ := s.store.Get(r, "session")

	if v, ok := session.Values["state"].(string); ok {
		return v
	}

	return ""
}

// Choose allows you to switch between two handlers depending on whether the
// expected user is signed in or not.
func (s *scopedSessions) WithToken(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, ok := s.get(r); ok && user.Me == s.me {
			ctx := context.WithValue(r.Context(), "user", user)
			h.ServeHTTP(w, r.WithContext(ctx))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}

// Shield will let the request continue if the expected user is signed in,
// otherwise they will be shown the DefaultSignedOut handler.
func (s *scopedSessions) Shield(signedIn http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user, ok := s.get(r); ok && user.Me == s.me {
			ctx := context.WithValue(r.Context(), "user", user)
			signedIn.ServeHTTP(w, r.WithContext(ctx))
		} else {
			s.defaultSignedOut.ServeHTTP(w, r)
		}
	})
}

// SignIn should be assigned to a route like /sign-in, it redirects users to the
// correct endpoint.
func (s *scopedSessions) SignIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := s.setState(w, r)
		if err != nil {
			http.Error(w, "could not start auth", http.StatusInternalServerError)
			return
		}

		redirectURL := s.auth.RedirectURL(s.ends, s.me, state)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

// Callback should be assigned to the redirectURL you configured for indieauth.
func (s *scopedSessions) Callback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state := s.getState(r)

		if r.FormValue("state") != state {
			http.Error(w, "state is bad", http.StatusBadRequest)
			return
		}

		token, err := s.auth.Exchange(s.ends, r.FormValue("code"), s.me)
		if err != nil || token.Me != s.me {
			http.Error(w, "nope", http.StatusForbidden)
			return
		}

		s.set(w, r, token)
		http.Redirect(w, r, s.Root, http.StatusFound)
	}
}

// SignOut will remove the session cookie for the user.
func (s *scopedSessions) SignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.set(w, r, indieauth.Token{})
		http.Redirect(w, r, s.Root, http.StatusFound)
	}
}
