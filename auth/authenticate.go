// Package auth implements a middleware providing indieauth.
//
// See the specification https://www.w3.org/TR/indieauth/.
package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"hawx.me/code/indieauth"
)

// Only delegates handling the request to next only if the user specified by me
// has provided authentication as expected by IndieAuth, either:
//
//   - passing a valid token as the 'access_token' form parameter, or
//   - including a valid token in the Authorization header with a prefix of
//     'Bearer'.
func Only(me string, next http.Handler) http.HandlerFunc {
	endpoints, err := indieauth.FindEndpoints(me)
	if err != nil {
		log.Fatal("ERR find-indieauth-endpoints;", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || strings.TrimSpace(auth) == "Bearer" {
			if r.FormValue("access_token") == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			auth = "Bearer " + r.FormValue("access_token")
		}

		req, err := http.NewRequest("GET", endpoints.Token.String(), nil)
		if err != nil {
			log.Println("ERR auth-make-request-failed;", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		req.Header.Add("Authorization", auth)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("ERR auth-request-failed;", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenData struct {
			Me       string `json:"me"`
			ClientID string `json:"client_id"`
			Scope    string `json:"scope"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
			log.Println("ERR auth-decode-token;", err)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}

		if tokenData.Me != me {
			log.Printf("ERR token-is-forbidden me=%s\n", tokenData.Me)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r.WithContext(
			context.WithValue(context.WithValue(r.Context(),
				scopesKey, strings.Fields(tokenData.Scope)),
				clientKey, tokenData.ClientID,
			),
		))
	}
}

const scopesKey = "__hawx.me/code/tally-ho:Scopes__"
const clientKey = "__hawx.me/code/tally-ho:ClientID__"

// HasScope checks that a request, authenticated with Only, contains one of the
// listed valid scopes.
func HasScope(w http.ResponseWriter, r *http.Request, valid ...string) bool {
	rv := r.Context().Value(scopesKey)
	if rv == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"insufficient_scope"}`, http.StatusUnauthorized)
		return false
	}

	scopes := rv.([]string)

	hasScope := intersects(valid, scopes)
	if !hasScope {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"insufficient_scope"}`, http.StatusUnauthorized)
		return false
	}

	return true
}

// ClientID returns the clientId that was issued for the token in a request that
// has been authenticated with Only.
func ClientID(r *http.Request) string {
	rv := r.Context().Value(clientKey)
	if rv == nil {
		return ""
	}

	return rv.(string)
}

func intersects(needles []string, list []string) bool {
	for _, needle := range needles {
		for _, item := range list {
			if item == needle {
				return true
			}
		}
	}

	return false
}
