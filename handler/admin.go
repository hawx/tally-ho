package handler

import "net/http"

func Admin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
<a href="/admin/sign-in">Sign-in</a>
`))
	}
}
