package handler

import (
	"log"
	"net/http"

	"hawx.me/code/tally-ho/blog"
)

func Admin(blog *blog.Blog, adminURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentPage, _ := blog.CurrentPage()
		user, ok := r.Context().Value("user").(userSession)

		if err := blog.RenderAdmin(w, struct {
			SignedIn    bool
			CurrentPage string
			AccessToken string
			Micropub    string
			Media       string
			AdminURL    string
		}{
			SignedIn:    ok,
			CurrentPage: currentPage,
			AccessToken: user.AccessToken,
			Micropub:    user.Micropub,
			Media:       user.Media,
			AdminURL:    adminURL,
		}); err != nil {
			log.Println(err)
		}
	}
}
