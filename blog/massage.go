package blog

import (
	"log"
	"net/url"
	"regexp"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/tally-ho/internal/mfutil"
	"mvdan.cc/xurls/v2"
)

var citeable = map[string]string{
	"like":     "like-of",
	"reply":    "in-reply-to",
	"bookmark": "bookmark-of",
}

// massage will do all of the magic to the data to make it nicer. It should be
// safe to call this when updating a post, so it should NOT overwrite any
// existing data.
func (b *Blog) massage(data map[string][]interface{}) {
	uid := uuid.New().String()

	relativeURL, _ := url.Parse("/entry/" + uid)
	location := b.config.BaseURL.ResolveReference(relativeURL).String()

	if len(data["uid"]) == 0 {
		data["uid"] = []interface{}{uid}
	}
	if len(data["url"]) == 0 {
		data["url"] = []interface{}{location}
	}

	if len(data["published"]) == 0 {
		data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}
	}

	kind := postTypeDiscovery(data)

	for k, v := range citeable {
		if kind == k {
			// safe because it only attempts to find cites for things that are strings
			s, ok := data[v][0].(string)
			if !ok {
				continue
			}

			cite, err := b.resolveCite(s)
			if err != nil {
				log.Printf("WARN get-cite; %v\n", err)
				continue
			}

			data[v] = []interface{}{cite}
		}
	}

	// kind could be changed by an update, so this is fine
	data["hx-kind"] = []interface{}{kind}

	if content, ok := data["content"]; ok && len(content) > 0 {
		// safe because it only attempts to autolink when content is a string
		if s, ok := content[0].(string); ok {
			reg := xurls.Strict()

			people := map[string][]string{}

			html := regexp.MustCompile("@?"+reg.String()).ReplaceAllStringFunc(s, func(u string) string {
				if u[0] == '@' {
					person, err := b.resolveCard(u[1:])
					if err != nil {
						log.Println("WARN get-person;", err)
					}
					if person != nil {
						if me, ok := person["me"].([]string); ok {
							people[u[1:]] = me
						}
						return `<a href="` + mfutil.Get(person, "properties.url").(string) + `">` + mfutil.Get(person, "properties.name", "properties.url").(string) + `</a>`
					}
				}

				return `<a href="` + u + `">` + u + `</a>`
			})

			data["content"] = []interface{}{map[string]interface{}{
				"text": s,
				"html": html,
			}}
			data["hx-people"] = []interface{}{people}
		}
	}
}

func postTypeDiscovery(data map[string][]interface{}) string {
	if rsvp, ok := data["rsvp"]; ok && len(rsvp) > 0 && (rsvp[0] == "yes" || rsvp[0] == "no" || rsvp[0] == "maybe") {
		return "rsvp"
	}

	if u, ok := data["in-reply-to"]; ok && len(u) > 0 {
		return "reply"
	}

	if u, ok := data["repost-of"]; ok && len(u) > 0 {
		return "repost"
	}

	if u, ok := data["like-of"]; ok && len(u) > 0 {
		return "like"
	}

	if u, ok := data["bookmark-of"]; ok && len(u) > 0 {
		return "bookmark"
	}

	if u, ok := data["video"]; ok && len(u) > 0 {
		return "video"
	}

	if u, ok := data["photo"]; ok && len(u) > 0 {
		return "photo"
	}

	if u, ok := data["read-of"]; ok && len(u) > 0 {
		return "read"
	}

	if u, ok := data["drank"]; ok && len(u) > 0 {
		return "drank"
	}

	if u, ok := data["checkin"]; ok && len(u) > 0 {
		return "checkin"
	}

	// I know the algorithm https://indieweb.org/post-type-discovery does more
	// than this, that is for another time
	if n, ok := data["name"]; ok && len(n) > 0 {
		return "article"
	}

	return "note"
}
