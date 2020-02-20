package blog

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"hawx.me/code/tally-ho/internal/mfutil"
	"mvdan.cc/xurls/v2"
	"willnorris.com/go/microformats"
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
		cite, err := getCite(data["like-of"][0].(string))
		if err != nil {
			log.Printf("WARN get-cite; %v\n", err)
		}
		data["like-of"] = []interface{}{cite}
		log.Printf("WARN get-cite; setting to '%s'\n", cite)
	}
	if kind == "reply" {
		cite, err := getCite(data["in-reply-to"][0].(string))
		if err != nil {
			log.Printf("WARN get-cite; %v\n", err)
		}
		data["in-reply-to"] = []interface{}{cite}
		log.Printf("WARN get-cite; setting to '%s'\n", cite)
	}
	if kind == "bookmark" {
		cite, err := getCite(data["bookmark-of"][0].(string))
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

	return location, nil
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

func getCite(u string) (cite map[string]interface{}, err error) {
	cite = map[string]interface{}{
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"url": {u},
		},
	}

	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	root, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	uURL, _ := url.Parse(u)
	data := microformats.ParseNode(root, uURL)

	for _, item := range data.Items {
		if contains("h-entry", item.Type) {
			props := map[string][]interface{}{
				"url": {u},
			}

			if names := item.Properties["name"]; len(names) > 0 {
				props["name"] = []interface{}{names[0]}
			}

			if authors := item.Properties["author"]; len(authors) > 0 {
				if author, ok := authors[0].(*microformats.Microformat); ok && contains("h-card", author.Type) {
					props["author"] = []interface{}{
						map[string]interface{}{
							"type":       []interface{}{"h-card"},
							"properties": author.Properties,
						},
					}
				}

				cite = map[string]interface{}{
					"type":       []interface{}{"h-cite"},
					"properties": props,
				}
				return
			}
		}
	}

	metas := htmlutil.SearchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.DataAtom == atom.Meta
	})

	var ogMeta struct {
		ok                    bool
		site_name, title, url string
	}
	ogMeta.url = u

	for _, meta := range metas {
		if htmlutil.Has(meta, "property") {
			switch htmlutil.Attr(meta, "property") {
			case "og:type":
				if content := htmlutil.Attr(meta, "content"); content == "article" {
					ogMeta.ok = true
				}
			case "og:site_name":
				ogMeta.site_name = htmlutil.Attr(meta, "content")
			case "og:title":
				ogMeta.title = htmlutil.Attr(meta, "content")
			case "og:url":
				ogMeta.url = htmlutil.Attr(meta, "content")
			}
		}
	}

	if ogMeta.ok && ogMeta.title != "" {
		props := map[string][]interface{}{
			"name": {ogMeta.title},
			"url":  {ogMeta.url},
		}

		if ogMeta.site_name != "" {
			props["author"] = []interface{}{
				map[string]interface{}{
					"type": []interface{}{"h-card"},
					"properties": map[string][]interface{}{
						"name": {ogMeta.site_name},
					},
				},
			}
		}

		cite = map[string]interface{}{
			"type":       []interface{}{"h-cite"},
			"properties": props,
		}
		return
	}

	titles := htmlutil.SearchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.DataAtom == atom.Title
	})

	if len(titles) > 0 {
		cite = map[string]interface{}{
			"type": []interface{}{"h-cite"},
			"properties": map[string][]interface{}{
				"url":  {u},
				"name": {htmlutil.TextOf(titles[0])},
			},
		}
		return
	}

	return cite, ErrNoName
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
