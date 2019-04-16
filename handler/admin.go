package handler

import (
	"net/http"

	"hawx.me/code/tally-ho/blog"
)

func Admin(blog *blog.Blog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentPage, _ := blog.CurrentPage()
		accessToken, ok := r.Context().Value("access_token").(string)

		blog.RenderAdmin(w, struct {
			SignedIn    bool
			CurrentPage string
			AccessToken string
			Micropub    string
		}{
			SignedIn:    ok,
			CurrentPage: currentPage,
			AccessToken: accessToken,
			Micropub:    "/micropub",
		})
	}
}
