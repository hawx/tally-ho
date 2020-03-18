package blog

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"willnorris.com/go/microformats"
)

var ErrNoName = errors.New("no name to find")

type Citer interface {
	Cite(string) (map[string]interface{}, error)
}

func (b *Blog) getCite(u string) (map[string]interface{}, error) {
	for _, citer := range b.citers {
		cite, err := citer.Cite(u)
		if err != nil {
			log.Printf("ERR get-cite url=%s; %v\n", u, err)
			return nil, nil
		}

		if cite == nil {
			continue
		}

		return cite, err
	}

	return getCite(u)
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
				props["name"] = names

				if contents := item.Properties["content"]; len(contents) > 0 {
					// check if a note
					if content, ok := contents[0].(map[string]interface{}); ok && content["value"] == props["name"][0] {
						if content["value"] == props["name"][0] {
							props["content"] = contents
							props["name"] = []interface{}{"a note"}
						}
					}
				}
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
