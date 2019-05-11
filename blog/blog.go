// Package blog writes to disk a blog, using data from different sources.
package blog

import (
	"html/template"
	"log"
	"os"
	"path/filepath"

	"hawx.me/code/tally-ho/micropub"
	"hawx.me/code/tally-ho/webmention"
	"hawx.me/code/tally-ho/writer"
)

type Looper struct {
	Blog *Blog
}

func (l *Looper) PostChanged(url string) error {
	return l.Blog.PostChanged(url)
}

type Blog struct {
	FileWriter writer.FileWriter
	Entries    *micropub.Reader
	Mentions   *webmention.Reader
	Templates  *template.Template
}

// PostChanged will render the post at the given url, and also render the page
// that the post belongs to.
func (b *Blog) PostChanged(url string) error {
	post, err := b.Post(url)
	if err != nil {
		return err
	}
	if err := b.RenderPost(post); err != nil {
		return err
	}

	page, err := b.Page(post.PageURL)
	if err != nil {
		return err
	}
	if err := b.RenderPage(page); err != nil {
		return err
	}

	return nil
}

// RenderPost writes a post.
func (b *Blog) RenderPost(post *Post) error {
	url := post.Properties["url"][0].(string)

	return b.write(url, "post.gotmpl", post)
}

// RenderPage writes the page, if it IsRoot then the root page will also be
// written. To ensure the next page link works it will write the previous page
// if there is only one post.
func (b *Blog) RenderPage(page *Page) error {
	if err := b.write(page.URL, "page.gotmpl", page); err != nil {
		return err
	}
	if page.IsRoot {
		if err := b.write(b.FileWriter.URL("/"), "page.gotmpl", page); err != nil {
			return err
		}
	}

	if len(page.Posts) == 1 {
		prev, err := page.Prev(b)
		if err != nil && err != ErrNoPage {
			return err
		}

		if prev != nil {
			// Don't use Render because if the previous page only had 1 post we'll start
			// doing everything...
			if err := b.write(prev.URL, "page.gotmpl", prev); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Blog) write(url, tmpl string, data interface{}) error {
	path := b.FileWriter.Path(url)
	dir := filepath.Dir(path)

	log.Println("mkdir", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	log.Println("writing", path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return b.Templates.ExecuteTemplate(file, tmpl, data)
}
