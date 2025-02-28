// Package webmention implements a handler for receiving webmentions.
//
// See the specification https://www.w3.org/TR/webmention/.
package webmention

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
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
			slog.Info("received webmention", slog.String("target", mention.target), slog.String("source", mention.source))

			if err := processMention(mention, blog); err != nil {
				slog.Error("process mention", slog.Any("err", err))
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			source = r.FormValue("source")
			target = r.FormValue("target")
		)

		if source == "" || target == "" {
			slog.Error("invalid webmention missing argument", slog.String("source", source), slog.String("target", target))
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(target, baseURL) {
			slog.Error("invalid webmention incorrect base url", slog.String("target", target))
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		slog.Info("webmention queued", slog.String("source", source), slog.String("target", target))
		mentions <- webmention{source: source, target: target}
		w.WriteHeader(http.StatusAccepted)
	}
}

func processMention(mention webmention, blog Blog) error {
	_, err := blog.Entry(mention.target)
	if err != nil {
		return errors.New("no such post at 'target'")
	}

	source, err := url.Parse(mention.source)
	if err != nil {
		return errors.New("could not parse 'source'")
	}

	// use a http client with a shortish timeout or this could block
	resp, err := http.Get(mention.source)
	if err != nil {
		return errors.New("could not retrieve 'source'")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone {
		if err := blog.Mention(mention.source, map[string][]interface{}{
			"hx-target": {mention.target},
			"hx-gone":   {true},
		}); err != nil {
			return fmt.Errorf("could not tombstone webmention: %w", err)
		}

		return nil
	}

	data := authorship.Parse(resp.Body, source)

	properties := map[string][]interface{}{}
	for _, item := range data.Items {
		if slices.Contains(item.Type, "h-entry") {
			properties = item.Properties
			break
		}
	}
	properties["hx-target"] = []interface{}{mention.target}

	if err := blog.Mention(mention.source, properties); err != nil {
		return fmt.Errorf("could not add webmention: %w", err)
	}

	return nil
}
