package admin

import (
	"html/template"
	"log"
	"net/http"

	"hawx.me/code/indieauth"
	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/micropub"
)

func Endpoint(
	adminURL, me, secret, webPath string,
	mr *micropub.Reader,
	templates *template.Template,
) (h http.Handler, err error) {
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
		"GET": session.WithToken(Admin(templates, mr, adminURL)),
	})

	router.HandleFunc("/sign-in", session.SignIn())
	router.HandleFunc("/callback", session.Callback())
	router.HandleFunc("/sign-out", session.SignOut())

	router.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir(webPath+"/static"))))

	return router, nil
}

func Admin(templates *template.Template, mr *micropub.Reader, adminURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentPage, _ := mr.CurrentPage()
		user, ok := r.Context().Value("user").(userSession)

		if err := templates.ExecuteTemplate(w, "admin.gotmpl", struct {
			SignedIn    bool
			CurrentPage string
			AccessToken string
			Micropub    string
			Media       string
			AdminURL    string
		}{
			SignedIn:    ok,
			CurrentPage: currentPage.Name,
			AccessToken: user.AccessToken,
			Micropub:    user.Micropub,
			Media:       user.Media,
			AdminURL:    adminURL,
		}); err != nil {
			log.Println(err)
		}
	}
}
