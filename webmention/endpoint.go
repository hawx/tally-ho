// Package webmention implements a handler for receiving webmentions.
//
// See the specification https://www.w3.org/TR/webmention/.
package webmention

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"

	"hawx.me/code/mux"
	"hawx.me/code/numbersix"
	"willnorris.com/go/microformats"
)

type mentionBlog interface {
	PostByURL(url string) (map[string][]interface{}, error)
}

type webmention struct {
	source, target string
}

func Endpoint(db *sql.DB, blog mentionBlog) (h http.Handler, err error) {
	mentions, err := numbersix.For(db, "mentions")
	if err != nil {
		return nil, err
	}

	return mux.Method{"POST": postHandler(mentions, blog)}, nil
}

func postHandler(db *numbersix.DB, blog mentionBlog) http.HandlerFunc {
	mentions := make(chan webmention, 100)

	go func() {
		for mention := range mentions {
			log.Println("got mention", mention.target, mention.source)

			if err := processMention(mention, blog, db); err != nil {
				log.Println(err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			source = r.FormValue("source")
			target = r.FormValue("target")
		)

		log.Println("queuing", source, target)
		if source == "" || target == "" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		mentions <- webmention{source: source, target: target}
		w.WriteHeader(http.StatusAccepted)
	}
}

func processMention(mention webmention, blog mentionBlog, db *numbersix.DB) error {
	_, err := blog.PostByURL(mention.target)
	if err != nil {
		return errors.New("  no such post at 'target'")
	}

	source, err := url.Parse(mention.source)
	if err != nil {
		return errors.New("  could not parse 'source'")
	}

	// use a http client with a shortish timeout or this could block
	resp, err := http.Get(mention.source)
	if err != nil {
		return errors.New("  could not retrieve 'source'")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		if err := upsertMention(db, mention.source, map[string][]interface{}{
			"hx-target": {mention.target},
			"gone":      {true},
		}); err != nil {
			return errors.New("  could not tombstone webmention: " + err.Error())
		}

		return nil
	}

	data := microformats.Parse(resp.Body, source)

	properties := map[string][]interface{}{}
	for _, item := range data.Items {
		if contains("h-entry", item.Type) {
			properties = item.Properties
			break
		}
	}
	properties["hx-target"] = []interface{}{mention.target}

	if err := upsertMention(db, mention.source, properties); err != nil {
		return errors.New("  could not add webmention: " + err.Error())
	}

	return nil
}

func contains(needle string, list []string) bool {
	for _, item := range list {
		if item == needle {
			return true
		}
	}

	return false
}
