package blog

import (
	"log"
	"net/url"
	"strings"
	"time"

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
