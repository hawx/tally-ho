package micropub

import (
	"net/http"

	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/blog"
)

func Endpoint(me string, blog *blog.Blog, mediaUploadURL string) (http.Handler, error) {
	return Authenticate(me, "create", mux.Method{
		"POST": postHandler(blog),
		"GET":  getHandler(blog, mediaUploadURL),
	}), nil
}
