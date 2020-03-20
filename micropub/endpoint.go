// Package micropub implements a micropub handler.
//
// See the specification https://www.w3.org/TR/micropub/.
package micropub

import (
	"net/http"

	"hawx.me/code/mux"
	"hawx.me/code/tally-ho/auth"
	"hawx.me/code/tally-ho/media"
)

type DB interface {
	Entry(url string) (data map[string][]interface{}, err error)
	Create(data map[string][]interface{}) (string, error)
	Update(url string, replace, add, delete map[string][]interface{}, deleteAlls []string) error
	Delete(url string) error
	Undelete(url string) error
}

// Endpoint returns a http.Handler exposing micropub. Only tokens issued for
// 'me' are allowed access to post or retrieve configuration.
func Endpoint(
	db DB,
	me string,
	mediaUploadURL string,
	syndicateTo []SyndicateTo,
	fw media.FileWriter,
) http.Handler {
	return auth.Only(me, mux.Method{
		"POST": postHandler(db, fw),
		"GET":  getHandler(db, mediaUploadURL, syndicateTo),
	})
}
