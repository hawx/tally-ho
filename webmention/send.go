package webmention

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/tomnomnom/linkheader"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

	links := searchAll(root, func(node *html.Node) bool {
		return node.Type == html.ElementNode &&
			(node.DataAtom == atom.Link || node.DataAtom == atom.A) &&
			hasAttr(node, "rel", "webmention") &&
			has(node, "href")
	})

	if len(links) > 0 {
		webmentionLinkURL, err := url.Parse(getAttr(links[0], "href"))
		if err != nil {
			return "", err
		}
		return targetURL.ResolveReference(webmentionLinkURL).String(), nil
	}

	return "", errors.New("nope")
}

func searchAll(node *html.Node, pred func(*html.Node) bool) (results []*html.Node) {
	if pred(node) {
		results = append(results, node)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result := searchAll(child, pred)
		if len(result) > 0 {
			results = append(results, result...)
		}
	}

	return
}

func hasAttr(node *html.Node, attrName, attrValue string) bool {
	values := strings.Fields(getAttr(node, attrName))

	for _, v := range values {
		if v == attrValue {
			return true
		}
	}

	return false
}

func has(node *html.Node, attrName string) bool {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return true
		}
	}

	return false
}

func getAttr(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
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
