package blog

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"hawx.me/code/tally-ho/internal/mfutil"
	"mvdan.cc/xurls/v2"
)

func (b *Blog) Create(data map[string][]interface{}) (string, error) {
	uid := uuid.New().String()

	relativeURL, _ := url.Parse("/entry/" + uid)
	location := b.Config.BaseURL.ResolveReference(relativeURL).String()

	data["uid"] = []interface{}{uid}
	data["url"] = []interface{}{location}

	if len(data["published"]) == 0 {
		data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}
	}

	kind := postTypeDiscovery(data)

	if kind == "like" {
		cite, err := b.getCite(data["like-of"][0].(string))
		if err != nil {
			log.Printf("WARN get-cite; %v\n", err)
		}
		data["like-of"] = []interface{}{cite}
		log.Printf("WARN get-cite; setting to '%s'\n", cite)
	}
	if kind == "reply" {
		cite, err := b.getCite(data["in-reply-to"][0].(string))
		if err != nil {
			log.Printf("WARN get-cite; %v\n", err)
		}
		data["in-reply-to"] = []interface{}{cite}
		log.Printf("WARN get-cite; setting to '%s'\n", cite)
	}
	if kind == "bookmark" {
		cite, err := b.getCite(data["bookmark-of"][0].(string))
		if err != nil {
			log.Printf("WARN get-cite; %v\n", err)
		}
		data["bookmark-of"] = []interface{}{cite}
		log.Printf("WARN get-cite; setting to '%s'\n", cite)
	}

	data["hx-kind"] = []interface{}{kind}

	if content, ok := data["content"]; ok && len(content) > 0 {
		if s, ok := content[0].(string); ok {
			html := xurls.Strict().ReplaceAllStringFunc(s, func(u string) string {
				return `<a href="` + u + `">` + u + `</a>`
			})

			data["content"] = []interface{}{map[string]interface{}{
				"text": s,
				"html": html,
			}}
		}
	}

	if err := b.entries.SetProperties(uid, data); err != nil {
		return location, err
	}

	go b.syndicate(location, data)
	go b.sendWebmentions(location, data)
	go b.hubPublish()

	return location, nil
}

func findAs(data map[string][]interface{}) []string {
	content, ok := mfutil.SafeGet(data, "content.html")
	if !ok {
		return []string{}
	}

	htmlContent, ok := content.(string)
	if !ok {
		return []string{}
	}

	root, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		log.Println("ERR find-as;", err)
		return []string{}
	}

	as := htmlutil.SearchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode &&
			node.DataAtom == atom.A &&
			htmlutil.Has(node, "href")
	})

	var links []string
	for _, a := range as {
		if val := htmlutil.Attr(a, "href"); val != "" {
			links = append(links, val)
		}
	}

	return links
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

var ErrNoName = errors.New("no name to find")

func contains(needle string, list []string) bool {
	for _, x := range list {
		if x == needle {
			return true
		}
	}
	return false
}
