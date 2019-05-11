package blog

import (
	"database/sql"
	"errors"
	"strings"
)

// ErrNoPage is returned as an error if a page with the URL, or a previous page,
// does not exist.
var ErrNoPage = errors.New("there is no such page")

func (b *Blog) Page(url string) (*Page, error) {
	parts := strings.SplitAfter(url, "/")
	baseURL := strings.Join(parts[:len(parts)-2], "")

	p, err := b.Entries.Page(url)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoPage
		}
		return nil, err
	}

	current, err := b.Entries.CurrentPage()
	if err != nil {
		return nil, err
	}

	entries, err := b.Entries.Entries(p.URL)
	if err != nil {
		return nil, err
	}

	var posts []map[string][]interface{}
	for _, entry := range entries {
		properties := entry.Properties
		posts = append(posts, properties)
	}

	var nextPage *PageRef
	next, err := b.Entries.PageAfter(p.URL)
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
	prev, err := b.Entries.PageBefore(p.URL)
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

	// URL the page is located at.
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

// Prev returns the previous, older page, if such a page exists.
func (p *Page) Prev(b *Blog) (*Page, error) {
	prev, err := b.Entries.PageBefore(p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoPage
		}
		return nil, err
	}

	return b.Page(prev.URL)
}
