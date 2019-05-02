// Package micropub implements a micropub handler.
//
// See the specification https://www.w3.org/TR/micropub/.
package micropub

import (
	"database/sql"
	"net/http"

	"hawx.me/code/mux"
	"hawx.me/code/numbersix"
	"hawx.me/code/tally-ho/writer"
)

type Notifier interface {
	PostChanged(url string) error
}

// Endpoint returns a http.Handler exposing micropub. Only tokens issued for
// 'me' are allowed access to post or retrieve configuration.
func Endpoint(
	db *sql.DB,
	me string,
	blog Notifier,
	mediaUploadURL string,
	uf writer.URLFactory,
) (h http.Handler, r *Reader, err error) {
	if err := migrate(db); err != nil {
		return nil, nil, err
	}

	entries, err := numbersix.For(db, "entries")
	if err != nil {
		return nil, nil, err
	}

	mdb := &micropubDB{
		entries: entries,
		sql:     db,
	}

	return authenticate(me, "create", mux.Method{
		"POST": postHandler(blog, uf, mdb),
		"GET":  getHandler(mdb, mediaUploadURL),
	}), &Reader{mdb}, nil
}
