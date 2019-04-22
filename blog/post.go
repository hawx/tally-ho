package blog

import (
	"html/template"
)

func (page *Page) Post(properties map[string][]interface{}) (*Post, error) {
	return &Post{
		Properties: properties,
		BaseURL:    page.BaseURL,
		PageURL:    page.URL,
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

func (p *Post) Render(tmpl *template.Template, w writer) error {
	url := p.Properties["url"][0].(string)

	return w.writePost(url, p)
}
