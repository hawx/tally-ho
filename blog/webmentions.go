package blog

import (
	"log"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/tally-ho/internal/htmlutil"
	"hawx.me/code/tally-ho/internal/mfutil"
	"hawx.me/code/tally-ho/webmention"
)

func (b *Blog) sendWebmentions(location string, data map[string][]interface{}) {
	// ensure that the entry exists
	time.Sleep(time.Second)

	links := findMentionedLinks(data)
	log.Printf("INFO sending-webmentions; %v\n", links)

	for _, link := range links {
		if err := webmention.Send(location, link); err != nil {
			log.Printf("ERR send-webmention source=%s target=%s; %v\n", location, link, err)
		}
	}
}

func (b *Blog) sendUpdateWebmentions(location string, oldData, newData map[string][]interface{}) {
	links := findMentionedLinks(newData)

	for _, oldLink := range findMentionedLinks(oldData) {
		if !contains(oldLink, links) {
			links = append(links, oldLink)
		}
	}

	log.Printf("INFO sending-webmentions; %v\n", links)

	for _, link := range links {
		if err := webmention.Send(location, link); err != nil {
			log.Printf("ERR send-webmention source=%s target=%s; %v\n", location, link, err)
		}
	}
}

func findMentionedLinks(data map[string][]interface{}) []string {
	var links []string

	links = append(links, findAs(data)...)

	for key, value := range data {
		if strings.HasPrefix(key, "hx-") ||
			strings.HasPrefix(key, "mp-") ||
			key == "url" ||
			len(value) == 0 {
			continue
		}

		if v, ok := mfutil.Get(data, key+".properties.url", key).(string); ok {
			if u, err := url.Parse(v); err == nil && u.IsAbs() {
				links = append(links, v)
			}
		}
	}

	return links
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
