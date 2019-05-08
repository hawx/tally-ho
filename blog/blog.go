package blog

import (
	"html/template"
	"regexp"
	"strings"

	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/writer"
)

type Blog struct {
	BaseURL    string
	FileWriter writer.FileWriter
	Store      *micropub.Reader
	Templates  *template.Template
}

// PostChanged will render the post at the given url, and also render the page
// that the post belongs to.
func (b *Blog) PostChanged(url string) error {
	post, err := FindPostByURL(url, b.Store)
	if err != nil {
		return err
	}
	if err := post.Render(b.Templates, b); err != nil {
		return err
	}

	page, err := FindPageByURL(post.PageURL, b.Store)
	if err != nil {
		return err
	}
	if err := page.Render(b.Store, b.Templates, b); err != nil {
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
