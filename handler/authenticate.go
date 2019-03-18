package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"hawx.me/code/indieauth"
)

func Authenticate(me, scope string, next http.Handler) http.HandlerFunc {
	endpoints, err := indieauth.FindEndpoints(me)
	if err != nil {
		log.Fatal("indieauth find endpoints:", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		return
		auth := r.Header.Get("Authorization")
		if auth == "" {
			auth = "Bearer " + r.FormValue("access_token")
		}
		if auth == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		req, err := http.NewRequest("GET", endpoints.Token.String(), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		req.Header.Add("Authorization", auth)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var tokenData struct {
			Me       string `json:"me"`
			ClientID string `json:"client_id"`
			Scope    string `json:"scope"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tokenData); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if tokenData.Me != me {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		hasScope := contains(scope, strings.Fields(tokenData.Scope))
		if !hasScope {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

func contains(needle string, list []string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}

	return false
}
