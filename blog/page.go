package blog

import (
	"database/sql"
	"errors"
	"html/template"
	"strings"

	"hawx.me/code/tally-ho/micropub"
)

var ErrNoPage = errors.New("there is no such page")

func FindPageByURL(url string, store *micropub.Reader) (*Page, error) {
	parts := strings.SplitAfter(url, "/")
	baseURL := strings.Join(parts[:len(parts)-2], "")

	p, err := store.Page(url)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoPage
		}
		return nil, err
	}

	current, err := store.CurrentPage()
	if err != nil {
		return nil, err
	}

	entries, err := store.Entries(p.Name)
	if err != nil {
		return nil, err
	}

	var posts []map[string][]interface{}
	for _, entry := range entries {
		properties := entry.Properties
		posts = append(posts, properties)
	}

	var nextPage *PageRef
	next, err := store.PageAfter(p.URL)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		nextPage = &PageRef{
			Name: next.Name,
			URL:  next.URL,
		}
	}

	var prevPage *PageRef
	prev, err := store.PageBefore(p.URL)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if err == nil {
		prevPage = &PageRef{
			Name: prev.Name,
			URL:  prev.URL,
		}
	}

	return &Page{
		Name:     p.Name,
		BaseURL:  baseURL,
		URL:      p.URL,
		Posts:    posts,
		IsRoot:   p.Name == current.Name,
		NextPage: nextPage,
		PrevPage: prevPage,
	}, nil
}

// A Page on the blog. The difference to a "normal" blog is that pages are given
// names instead of numbers -- I am not a number! -- and don't contain a set
// number of posts. This is quite nice when statically rendering as adding a new
// post will in the worst case affect only 2 pages (3 if you count the root).
type Page struct {
	// Name of the page.
	Name string

	// BaseURL for the blog.
	BaseURL string

	// URL the page will be located at.
	URL string

	// Posts for the page, sorted descending by publish date.
	Posts []map[string][]interface{}

	// IsRoot will be true if this is the latest page.
	IsRoot bool

	// NextPage and PrevPage will hold references to the respective pages, if they don't
	// exist then they will be nil.
	NextPage, PrevPage *PageRef
}

// A PageRef is a simple reference to a page, used for the next/prev linking.
type PageRef struct {
	// Name of the page.
	Name string

	// URL the page will be located at.
	URL string
}

// Next returns the next, newer page, if such a page exists.
func (p *Page) Next(store *micropub.Reader) (*Page, error) {
	next, err := store.PageAfter(p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoPage
		}
		return nil, err
	}

	return FindPageByURL(next.URL, store)
}

// Prev returns the previous, older page, if such a page exists.
func (p *Page) Prev(store *micropub.Reader) (*Page, error) {
	prev, err := store.PageBefore(p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoPage
		}
		return nil, err
	}

	return FindPageByURL(prev.URL, store)
}

// Render writes the page, if it IsRoot then the root page will also be
// written. To ensure the next page link works it will write the previous page
// if there is only one post.
func (p *Page) Render(store *micropub.Reader, tmpl *template.Template, w writer2) error {
	if err := w.writePage(p.URL, p); err != nil {
		return err
	}
	if err := w.writeRoot(p); err != nil {
		return err
	}

	if len(p.Posts) == 1 {
		prev, err := p.Prev(store)
		if err != nil && err != ErrNoPage {
			return err
		}

		if prev != nil {
			// Don't use Render because if the previous page only had 1 post we'll start
			// doing everything...
			if err := w.writePage(prev.URL, prev); err != nil {
				return err
			}
		}
	}

	// render index.html somehow...
	return nil
}
