package blog

import (
	"bytes"
	"encoding/xml"
	"html/template"
)

type Feed struct {
	XMLName     xml.Name   `xml:"rss"`
	Version     string     `xml:"version,attr"`
	Title       string     `xml:"channel>title"`
	Description string     `xml:"channel>description"`
	Link        string     `xml:"channel>link"`
	Items       []FeedItem `xml:"channel>item"`
}

func (b *Blog) mapToFeed(page *Page) Feed {
	var items []FeedItem
	for _, entry := range page.Posts {
		items = append(items, mapToFeedItem(b.Templates, entry))
	}

	return Feed{
		Version:     "2.0",
		Title:       b.Meta.Title,
		Description: b.Meta.Description,
		Link:        page.BaseURL,
		Items:       items,
	}
}

type FeedItem struct {
	Title       string              `xml:"title,omitempty"`
	Description FeedItemDescription `xml:"description"`
	Link        string              `xml:"link"`
	PubDate     string              `xml:"pubDate"`
}

type FeedItemDescription struct {
	Text string `xml:",cdata"`
}

func mapToFeedItem(templates *template.Template, post map[string][]interface{}) FeedItem {
	var description bytes.Buffer
	templates.ExecuteTemplate(&description, "content.gotmpl", post)

	return FeedItem{
		Title:       templateGetOr(post, "name", "").(string),
		Description: FeedItemDescription{Text: description.String()},
		Link:        templateMustGet(post, "url").(string),
		PubDate:     templateMustGet(post, "published").(string),
	}
}
