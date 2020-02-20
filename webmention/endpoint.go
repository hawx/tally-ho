// Package webmention implements a handler for receiving webmentions.
//
// See the specification https://www.w3.org/TR/webmention/.
package webmention

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"hawx.me/code/microformats/authorship"
	"hawx.me/code/mux"
)

type Blog interface {
	Entry(url string) (data map[string][]interface{}, err error)
	Mention(source string, data map[string][]interface{}) error
	BaseURL() string
}

type webmention struct {
	source, target string
}

// Endpoint receives webmentions, immediately returning a response of Accepted,
// and processing them asynchronously.
func Endpoint(blog Blog) http.Handler {
	return mux.Method{"POST": postHandler(blog)}
}

func postHandler(blog Blog) http.HandlerFunc {
	mentions := make(chan webmention, 100)
	baseURL := blog.BaseURL()

	go func() {
		for mention := range mentions {
			log.Printf("INFO received-webmention target=%s source=%s\n", mention.target, mention.source)

			if err := processMention(mention, blog); err != nil {
				log.Println("ERR process-mention;", err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			source = r.FormValue("source")
			target = r.FormValue("target")
		)

		if source == "" || target == "" {
			log.Printf("ERR invalid-webmention source=%s target=%s; missing argument\n", source, target)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(target, baseURL) {
			log.Printf("ERR invalid-webmention target=%s; incorrect base url\n", target)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		log.Printf("INFO webmention-queued source=%s target=%s\n", source, target)
		mentions <- webmention{source: source, target: target}
		w.WriteHeader(http.StatusAccepted)
	}
}

func processMention(mention webmention, blog Blog) error {
	_, err := blog.Entry(mention.target)
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
		if err := blog.Mention(mention.source, map[string][]interface{}{
			"hx-target": {mention.target},
			"hx-gone":   {true},
		}); err != nil {
			return errors.New("  could not tombstone webmention: " + err.Error())
		}

		return nil
	}

	data := authorship.Parse(resp.Body, source)

	properties := map[string][]interface{}{}
	for _, item := range data.Items {
		if contains("h-entry", item.Type) {
			properties = item.Properties
			break
		}
	}
	properties["hx-target"] = []interface{}{mention.target}

	if err := blog.Mention(mention.source, properties); err != nil {
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
