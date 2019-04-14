package handler

import (
	"fmt"
	"log"
	"net/http"

	"hawx.me/code/tally-ho/blog"
)

func Admin(signedIn bool, blog *blog.Blog) http.HandlerFunc {
	if !signedIn {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`
<a href="/admin/sign-in">Sign-in</a>
`))
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		currentPage, _ := blog.CurrentPage()
		accessToken := r.Context().Value("access_token").(string)

		fmt.Fprintf(w, `
<h1>welcome</h1>

<p>current page is "%v"</p>

<form method="post" action="/admin/page">
  <label for="page">Page</label>
  <input name="page" id="page" />

  <button type="submit">Set next page</button>
</form>

<form method="post" action="/micropub">
  <label for="name">Name</label>
  <input name="name" id="name" />

  <label for="content">Content</label>
  <textarea name="content" id="content"></textarea>

  <input type="hidden" name="access_token" value="%v" />
  <input type="hidden" name="h" value="entry" />

  <button type="submit">Post</button>
</form>
`, currentPage, accessToken)
	}
}

func SetPage(blog *blog.Blog) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := blog.SetNextPage(r.FormValue("page")); err != nil {
			log.Println(err)
		}

		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}
