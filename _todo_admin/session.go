package admin

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"hawx.me/code/indieauth"
	"willnorris.com/go/microformats"
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

func micropubEndpoint(me string) (string, error) {
	meURL, err := url.Parse(me)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(me)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data := microformats.Parse(resp.Body, meURL)
	if len(data.Rels["micropub"]) == 0 {
		return "", errors.New("no micropub endpoint")
	}

	return data.Rels["micropub"][0], nil
}

func mediaEndpoint(micropub, token string) (string, error) {
	micropubURL, err := url.Parse(micropub)
	if err != nil {
		return "", err
	}

	query := micropubURL.Query()
	query.Add("q", "config")
	micropubURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", micropubURL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Media string `json:"media-endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.Media == "" {
		return "", errors.New("no media endpoint")
	}

	return data.Media, nil
}

func (s *scopedSessions) set(w http.ResponseWriter, r *http.Request, token indieauth.Token) {
	session, _ := s.store.Get(r, "session")

	micropub, err := micropubEndpoint(token.Me)
	if err != nil {
		log.Println(err)
		return
	}

	media, err := mediaEndpoint(micropub, token.AccessToken)
	if err != nil {
		log.Println(err)
		return
	}

	session.Values["user"] = userSession{
		Me:          token.Me,
		AccessToken: token.AccessToken,
		Micropub:    micropub,
		Media:       media,
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
