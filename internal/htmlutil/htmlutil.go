package htmlutil

import (
	"strings"

	"golang.org/x/net/html"
)

func SearchAll(node *html.Node, pred func(*html.Node) bool) (results []*html.Node) {
	if pred(node) {
		results = append(results, node)
		return
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		result := SearchAll(child, pred)
		if len(result) > 0 {
			results = append(results, result...)
		}
	}

	return
}

func HasAttr(node *html.Node, attrName, attrValue string) bool {
	values := strings.Fields(Attr(node, attrName))

	for _, v := range values {
		if v == attrValue {
			return true
		}
	}

	return false
}

func Has(node *html.Node, attrName string) bool {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return true
		}
	}

	return false
}

func Attr(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func TextOf(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	var parts []string

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		parts = append(parts, TextOf(c))
	}

	return strings.Join(parts, " ")
}
