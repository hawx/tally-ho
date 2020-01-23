// Package micropub implements a micropub handler.
//
// See the specification https://www.w3.org/TR/micropub/.
package micropub

import (
	"net/http"

	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/auth"
	"hawx.me/code/tally-ho/media"
	"hawx.me/code/tally-ho/syndicate"
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
	syndicators map[string]syndicate.Syndicator,
	fw media.FileWriter,
) http.Handler {
	return auth.Only(me, "create", mux.Method{
		"POST": postHandler(db, fw),
		"GET":  getHandler(db, mediaUploadURL, syndicators),
	})
}
