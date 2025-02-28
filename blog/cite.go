package blog

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"willnorris.com/go/microformats"
)

var ErrNoName = errors.New("no name to find")

type CiteResolver interface {
	ResolveCite(string) (map[string]any, error)
}

func (b *Blog) resolveCite(u string) (map[string]any, error) {
	for _, citer := range b.citeResolvers {
		cite, err := citer.ResolveCite(u)
		if err != nil {
			b.logger.Error("resolve cite", slog.String("url", u), slog.Any("err", err))
			return nil, nil
		}

		if cite == nil {
			continue
		}

		return cite, err
	}

	return resolveCite(u)
}

func resolveCite(u string) (cite map[string]any, err error) {
	cite = map[string]any{
		"type": []any{"h-cite"},
		"properties": map[string][]any{
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
			props := map[string][]any{
				"url": append([]any{u}, item.Properties["syndication"]...),
			}

			if names := item.Properties["name"]; len(names) > 0 {
				props["name"] = names

				if contents := item.Properties["content"]; len(contents) > 0 {
					// check if a note
					if content, ok := contents[0].(map[string]any); ok && content["value"] == props["name"][0] {
						if content["value"] == props["name"][0] {
							props["content"] = contents
							props["name"] = []any{"a note"}
						}
					}
				}
			}

			if authors := item.Properties["author"]; len(authors) > 0 {
				if author, ok := authors[0].(*microformats.Microformat); ok && contains("h-card", author.Type) {
					props["author"] = []any{
						map[string]any{
							"type":       []any{"h-card"},
							"properties": author.Properties,
						},
					}
				}

				cite = map[string]any{
					"type":       []any{"h-cite"},
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
		props := map[string][]any{
			"name": {ogMeta.title},
			"url":  {ogMeta.url},
		}

		if ogMeta.site_name != "" {
			props["author"] = []any{
				map[string]any{
					"type": []any{"h-card"},
					"properties": map[string][]any{
						"name": {ogMeta.site_name},
					},
				},
			}
		}

		cite = map[string]any{
			"type":       []any{"h-cite"},
			"properties": props,
		}
		return
	}

	titles := htmlutil.SearchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode && node.DataAtom == atom.Title
	})

	if len(titles) > 0 {
		cite = map[string]any{
			"type": []any{"h-cite"},
			"properties": map[string][]any{
				"url":  {u},
				"name": {htmlutil.TextOf(titles[0])},
			},
		}
		return
	}

	return cite, ErrNoName
}
