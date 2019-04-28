// Package micropub implements a micropub handler.
//
// See the specification https://www.w3.org/TR/micropub/.
package micropub

import (
	"net/http"

	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/blog"
)

// Endpoint returns a http.Handler exposing micropub. Only tokens issued for
// 'me' are allowed access to post or retrieve configuration.
func Endpoint(me string, blog *blog.Blog, mediaUploadURL string) (http.Handler, error) {
	return authenticate(me, "create", mux.Method{
		"POST": postHandler(blog),
		"GET":  getHandler(blog, mediaUploadURL),
	}), nil
}
