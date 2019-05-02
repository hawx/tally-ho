package blog

import (
	"html/template"
	"strings"

	"hawx.me/code/tally-ho/micropub"
)

func FindPostByURL(url string, store *micropub.Reader) (*Post, error) {
	parts := strings.SplitAfter(url, "/")
	baseURL := strings.Join(parts[:len(parts)-3], "")
	pageURL := strings.Join(parts[:len(parts)-2], "")

	post, err := store.Post(url)
	if err != nil {
		return nil, err
	}

	return &Post{
		Properties: post,
		BaseURL:    baseURL,
		PageURL:    pageURL,
	}, nil
}

type Post struct {
	// Properties of the post.
	Properties map[string][]interface{}

	// BaseURL for the blog.
	BaseURL string

	// PageURL is the full URL of the page the post belongs to.
	PageURL string
}

func (post *Post) Render(tmpl *template.Template, w writer2) error {
	url := post.Properties["url"][0].(string)

	return w.writePost(url, post)
}
