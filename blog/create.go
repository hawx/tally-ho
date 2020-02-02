package blog

import (
	"log"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"hawx.me/code/tally-ho/webmention"
	"mvdan.cc/xurls/v2"
)

func (b *Blog) Create(data map[string][]interface{}) (location string, err error) {
	kind := postTypeDiscovery(data)

	if kind == "like" {
		cite, err := getCite(data["like-of"][0].(string))
		if err != nil {
			log.Printf("WARN get-cite; %v\n", err)
		}
		data["like-of"] = []interface{}{cite}
		log.Printf("WARN get-cite; setting to '%s'\n", cite)
	}

	data["hx-kind"] = []interface{}{kind}

	if content, ok := data["content"]; ok && len(content) > 0 {
		if s, ok := content[0].(string); ok {
			html := xurls.Strict().ReplaceAllStringFunc(s, func(u string) string {
				return `<a href="` + u + `">` + u + `</a>`
			})

			data["content"] = []interface{}{map[string]string{
				"text": s,
				"html": html,
			}}
		}
	}

	relativeLocation, err := b.DB.Create(data)
	if err != nil {
		return
	}

	relativeURL, _ := url.Parse(relativeLocation)
	location = b.Config.BaseURL.ResolveReference(relativeURL).String()

	go b.syndicate(location, data)
	go b.sendWebmentions(location, data)

	return
}

func (b *Blog) syndicate(location string, data map[string][]interface{}) {
	if syndicateTos, ok := data["mp-syndicate-to"]; ok && len(syndicateTos) > 0 {
		for _, syndicateTo := range syndicateTos {
			if syndicator, ok := b.Syndicators[syndicateTo.(string)]; ok {
				syndicatedLocation, err := syndicator.Create(data)
				if err != nil {
					log.Printf("ERR syndication to=%s uid=%s; %v\n", syndicator.Name(), data["uid"][0], err)
					continue
				}

				if err := b.Update(location, empty, map[string][]interface{}{
					"syndication": {syndicatedLocation},
				}, empty, []string{}); err != nil {
					log.Printf("ERR confirming-syndication to=%s uid=%s; %v\n", syndicator.Name(), data["uid"][0], err)
				}
			}
		}
	}
}

func (b *Blog) sendWebmentions(location string, data map[string][]interface{}) {
	var links []string

	links = append(links, findAs(data)...)
	// and like-of
	// etc.

	for _, link := range links {
		if err := webmention.Send(location, link); err != nil {
			log.Printf("ERR send-webmention source=%s target=%s; %v\n", location, link, err)
		}
	}
}

func findAs(data map[string][]interface{}) []string {
	if content, ok := data["content"]; ok && len(content) > 0 {
		if contents, ok := content[0].(map[string]string); ok {
			if htmlContent, ok := contents["html"]; ok && len(htmlContent) > 0 {
				root, err := html.Parse(strings.NewReader(htmlContent))
				if err != nil {
					log.Println("ERR send-webmentions;", err)
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
		}
	}

	return []string{}
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
