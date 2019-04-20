package handler

import (
	"net/http"

	"hawx.me/code/indieauth"
	"hawx.me/code/tally-ho/blog"
)

func Admin(blog *blog.Blog, adminURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentPage, _ := blog.CurrentPage()
		token, ok := r.Context().Value("token").(indieauth.Token)

		blog.RenderAdmin(w, struct {
			SignedIn    bool
			CurrentPage string
			AccessToken string
			Micropub    string
			AdminURL    string
		}{
			SignedIn:    ok,
			CurrentPage: currentPage,
			AccessToken: token.AccessToken,
			Micropub:    "/micropub",
			AdminURL:    adminURL,
		})
	}
}
