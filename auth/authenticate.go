package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"hawx.me/code/indieauth"
)

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
			context.WithValue(r.Context(), scopesKey, strings.Fields(tokenData.Scope)),
		))
	}
}

const scopesKey = "__hawx.me/code/tally-ho:Scopes__"

func HasScope(w http.ResponseWriter, r *http.Request, scope string) bool {
	rv := r.Context().Value(scopesKey)
	if rv == nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"insufficient_scope"}`, http.StatusUnauthorized)
		return false
	}

	scopes := rv.([]string)

	hasScope := contains(scope, scopes)
	if !hasScope {
		log.Printf("ERR token-missing-scope wanted=%s; %s\n", scope, strings.Join(scopes, " "))
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"insufficient_scope"}`, http.StatusUnauthorized)
		return false
	}

	return true
}

func contains(needle string, list []string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}

	return false
}
