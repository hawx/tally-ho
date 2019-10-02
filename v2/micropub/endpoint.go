// Package micropub implements a micropub handler.
//
// See the specification https://www.w3.org/TR/micropub/.
package micropub

import (
	"database/sql"
	"net/http"

	"hawx.me/code/mux"
	"hawx.me/code/numbersix"
)

// Endpoint returns a http.Handler exposing micropub. Only tokens issued for
// 'me' are allowed access to post or retrieve configuration.
func Endpoint(
	db *sql.DB,
	me string,
	mediaUploadURL string,
) (h http.Handler, r *Reader, err error) {
	entries, err := numbersix.For(db, "entries")
	if err != nil {
		return nil, nil, err
	}

	mdb := &micropubDB{
		entries: entries,
		sql:     db,
	}

	return authenticate(me, "create", mux.Method{
		"POST": postHandler(mdb),
		"GET":  getHandler(mdb, mediaUploadURL),
	}), &Reader{mdb}, nil
}
