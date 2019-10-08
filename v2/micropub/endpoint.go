// Package micropub implements a micropub handler.
//
// See the specification https://www.w3.org/TR/micropub/.
package micropub

import (
	"net/http"

	"hawx.me/code/mux"
)

type DB interface {
	getDB
	postDB
}

// Endpoint returns a http.Handler exposing micropub. Only tokens issued for
// 'me' are allowed access to post or retrieve configuration.
func Endpoint(
	db DB,
	me string,
	mediaUploadURL string,
) http.Handler {
	return authenticate(me, "create", mux.Method{
		"POST": postHandler(db),
		"GET":  getHandler(db, mediaUploadURL),
	})
}
