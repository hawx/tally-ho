package handler

import (
	"log"
	"net/http"
	"net/url"

	"willnorris.com/go/microformats"
)

type mentionBlog interface {
	PostByURL(url string) (map[string][]interface{}, error)
	// MentionSourceAllowed will check if the source URL or host of the source URL
	// has been blacklisted.
	MentionSourceAllowed(url string) bool
	// AddMention will add the properties to a new webmention, or if a mention
	// already exists for the sourceURL update those properties.
	AddMention(sourceURL string, data map[string][]interface{}) error
}

type webmention struct {
	source, target string
}

func Mention(blog mentionBlog) http.HandlerFunc {
	mentions := make(chan webmention, 100)

	go func() {
		for mention := range mentions {
			log.Println("got mention", mention.target, mention.source)

			_, err := blog.PostByURL(mention.target)
			if err != nil {
				log.Println("  no such post at 'target'")
				continue
			}

			source, err := url.Parse(mention.source)
			if err != nil {
				log.Println("  could not parse 'source'")
				continue
			}

			// use a http client with a shortish timeout or this could block
			resp, err := http.Get(mention.source)
			if err != nil {
				log.Println("  could not retrieve 'source'")
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusGone {
				if err := blog.AddMention(mention.source, map[string][]interface{}{
					"hx-target": {mention.target},
					"gone":      {true},
				}); err != nil {
					log.Println("  could not tombstone webmention:", err)
				}

				continue
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
			if err := blog.AddMention(mention.source, properties); err != nil {
				log.Println("  could not add webmention:", err)
			}
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) {
		var (
			source = r.FormValue("source")
			target = r.FormValue("target")
		)

		if source == "" || target == "" {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		mentions <- webmention{source: source, target: target}
		w.WriteHeader(http.StatusAccepted)
	}
}
