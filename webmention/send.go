package webmention

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
)

func Send(source, target string) error {
	endpoint, err := discoverEndpoint(target)
	if err != nil {
		return err
	}

	resp, err := http.PostForm(endpoint, url.Values{
		"source": {source},
		"target": {target},
	})
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func discoverEndpoint(target string) (string, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return "", err
	}

	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	for _, header := range resp.Header["Link"] {
		link := hrefByRel("webmention", linkheader.Parse(header))

		if link == "" {
			continue
		}

		linkURL, err := url.Parse(link)
		if err != nil {
			return "", err
		}

		return targetURL.ResolveReference(linkURL).String(), nil
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	links := htmlutil.SearchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode &&
			(node.DataAtom == atom.Link || node.DataAtom == atom.A) &&
			htmlutil.HasAttr(node, "rel", "webmention") &&
			htmlutil.Has(node, "href")
	})

	if len(links) > 0 {
		webmentionLinkURL, err := url.Parse(htmlutil.Attr(links[0], "href"))
		if err != nil {
			return "", err
		}
		return targetURL.ResolveReference(webmentionLinkURL).String(), nil
	}

	return "", errors.New("nope")
}

func hrefByRel(rel string, links linkheader.Links) string {
	for _, link := range links {
		for _, r := range strings.Fields(link.Rel) {
			if r == rel {
				return link.URL
			}
		}
	}

	return ""
}
