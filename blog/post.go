package blog

import (
	"strings"

	"hawx.me/code/numbersix"
)

func (b *Blog) Post(url string) (*Post, error) {
	parts := strings.SplitAfter(url, "/")
	baseURL := strings.Join(parts[:len(parts)-3], "")
	pageURL := strings.Join(parts[:len(parts)-2], "")

	post, err := b.Entries.Post(url)
	if err != nil {
		return nil, err
	}

	reactions, err := b.Mentions.ForPost(url)
	if err != nil {
		return nil, err
	}

	return &Post{
		Meta:       b.Meta,
		BaseURL:    baseURL,
		PageURL:    pageURL,
		URL:        url,
		Properties: post,
		Reactions:  reactions,
	}, nil
}

type Post struct {
	// Meta data for the blog.
	Meta Meta

	// BaseURL for the blog.
	BaseURL string

	// PageURL is the full URL of the page the post belongs to.
	PageURL string

	// URL the post is located at.
	URL string

	// Properties of the post.
	Properties map[string][]interface{}

	// Reactions is a list of webmentions received for the post.
	Reactions []numbersix.Group
}
