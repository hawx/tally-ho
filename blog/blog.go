package blog

import (
	"html/template"
	"regexp"
	"strings"

	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/writer"
)

type Blog struct {
	fw        writer.FileWriter
	baseURL   string
	templates *template.Template
	Store     *micropub.Reader
}

type Options struct {
	Fw        writer.FileWriter
	BaseURL   string
	Templates *template.Template
	Reader    *micropub.Reader
}

func New(options Options) (*Blog, error) {
	return &Blog{
		fw:        options.Fw,
		baseURL:   options.BaseURL,
		templates: options.Templates,
		Store:     options.Reader,
	}, nil
}

// PostChanged will render the post at the given url, and also render the page
// that the post belongs to.
func (b *Blog) PostChanged(url string) error {
	post, err := FindPostByURL(url, b.Store)
	if err != nil {
		return err
	}
	if err := post.Render(b.templates, b); err != nil {
		return err
	}

	page, err := FindPageByURL(post.PageURL, b.Store)
	if err != nil {
		return err
	}
	if err := page.Render(b.Store, b.templates, b); err != nil {
		return err
	}

	return nil
}

var nonWord = regexp.MustCompile("\\W+")

func slugify(s string) string {
	s = strings.ReplaceAll(s, "'", "")
	s = nonWord.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	return s
}
