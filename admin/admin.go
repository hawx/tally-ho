package admin

import (
	"log"
	"net/http"

	"hawx.me/code/indieauth"
	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/blog"
)

func Endpoint(adminURL, me, secret, webPath string, blog *blog.Blog) (h http.Handler, err error) {
	auth, err := indieauth.Authorization(adminURL, adminURL+"callback", []string{"create"})
	if err != nil {
		log.Fatal(err)
	}

	session, err := NewScopedSessions(me, secret, auth)
	if err != nil {
		return
	}
	session.Root = adminURL

	router := http.NewServeMux()

	router.Handle("/", mux.Method{
		"GET": session.WithToken(Admin(blog, adminURL)),
	})

	router.HandleFunc("/sign-in", session.SignIn())
	router.HandleFunc("/callback", session.Callback())
	router.HandleFunc("/sign-out", session.SignOut())

	router.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir(webPath+"/static"))))

	return router, nil
}

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
