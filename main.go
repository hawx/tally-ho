package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"hawx.me/code/indieauth"
	"hawx.me/code/mux"
	"hawx.me/code/route"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/blog"
	"hawx.me/code/tally-ho/handler"
)

func main() {
	var (
		port      = flag.String("port", "8080", "")
		socket    = flag.String("socket", "", "")
		me        = flag.String("me", "", "")
		dbPath    = flag.String("db", "file::memory:", "")
		baseURL   = flag.String("base-url", "http://localhost:8080/", "")
		basePath  = flag.String("base-path", "/tmp/", "")
		mediaURL  = flag.String("media-url", "http://localhost:8080/_media/", "")
		mediaPath = flag.String("media-path", "/tmp/", "")
		adminURL  = flag.String("admin-url", "http://localhost:8080/admin/", "")
		webPath   = flag.String("web", "web", "")
		secret    = flag.String("secret", "", "")
	)
	flag.Parse()

	mediaWriter, err := blog.NewFileWriter(*mediaPath, *mediaURL)
	if err != nil {
		log.Println("creating mediawriter:", err)
		return
	}

	blog, err := blog.New(blog.Options{
		WebPath:  *webPath,
		BaseURL:  *baseURL,
		BasePath: *basePath,
		DbPath:   *dbPath,
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer blog.Close()

	if flag.NArg() == 1 && flag.Arg(0) == "render" {
		if err := blog.RenderAll(); err != nil {
			log.Fatal(err)
		}

		return
	}

	auth, err := indieauth.Authorization(*adminURL, *adminURL+"callback", []string{"create"})
	if err != nil {
		log.Fatal(err)
	}

	session, err := newScopedSessions(*me, *secret, auth)
	if err != nil {
		log.Fatal(err)
	}
	session.root = *adminURL

	if *me == "" {
		log.Fatal("--me must be provided")
	}

	route.HandleFunc("/admin/sign-in", session.SignIn())
	route.HandleFunc("/admin/callback", session.Callback())
	route.HandleFunc("/admin/sign-out", session.SignOut())

	route.Handle("/admin", mux.Method{
		"GET": session.WithToken(handler.Admin(blog, *adminURL)),
	})

	route.Handle("/micropub", handler.Authenticate(*me, "create", mux.Method{
		"POST": handler.Post(blog),
		"GET":  handler.Configuration(blog),
	}))

	route.Handle("/webmention", mux.Method{
		"POST": handler.Mention(blog),
	})

	route.Handle("/media", mux.Method{
		"POST": handler.Media(mediaWriter),
	})

	serve.Serve(*port, *socket, route.Default)
}

type scopedSessions struct {
	me               string
	store            sessions.Store
	auth             *indieauth.AuthorizationConfig
	ends             indieauth.Endpoints
	defaultSignedOut http.Handler
	root             string
}

func newScopedSessions(me, secret string, auth *indieauth.AuthorizationConfig) (*scopedSessions, error) {
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
		root:             "/",
	}, nil
}

func (s *scopedSessions) get(r *http.Request) (me, accessToken string) {
	session, _ := s.store.Get(r, "session")

	if me, ok := session.Values["me"].(string); ok {
		if accessToken, ok = session.Values["access_token"].(string); ok {
			return me, accessToken
		}
	}

	return "", ""
}

func (s *scopedSessions) set(w http.ResponseWriter, r *http.Request, token indieauth.Token) {
	session, _ := s.store.Get(r, "session")
	session.Values["me"] = token.Me
	session.Values["access_token"] = token.AccessToken
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
		if me, accessToken := s.get(r); me == s.me {
			ctx := context.WithValue(r.Context(), "access_token", accessToken)
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
		if me, accessToken := s.get(r); me == s.me {
			ctx := context.WithValue(r.Context(), "access_token", accessToken)
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
		http.Redirect(w, r, s.root, http.StatusFound)
	}
}

// SignOut will remove the session cookie for the user.
func (s *scopedSessions) SignOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.set(w, r, indieauth.Token{})
		http.Redirect(w, r, s.root, http.StatusFound)
	}
}
