package data

import (
	"database/sql"
	"testing"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"hawx.me/code/assert"
)

type fakeURLFactory struct{}

func (f fakeURLFactory) PostURL(pageURL, slug string) string {
	return ""
}

func TestPages(t *testing.T) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite3", "file::memory:")
	if !assert.Nil(err) {
		return
	}

	store, err := Open(db, fakeURLFactory{})
	if !assert.Nil(err) {
		return
	}
	defer store.Close()

	// initially no pages
	_, err = store.CurrentPage()
	assert.Equal(sql.ErrNoRows, err)

	// add a page
	assert.Nil(store.SetNextPage("what is this", "http://example.com/blog/what-is-this"))
	if page, err := store.CurrentPage(); assert.Nil(err) {
		assert.Equal(page.Name, "what is this")
		assert.Equal(page.URL, "http://example.com/blog/what-is-this")
	}

	// add a few more pages
	assert.Nil(store.SetNextPage("page no 2", "http://example.com/blog/2"))
	assert.Nil(store.SetNextPage("page no 3", "http://example.com/blog/3"))
	assert.Nil(store.SetNextPage("page no 4", "http://example.com/blog/4"))
	assert.Nil(store.SetNextPage("page no 5", "http://example.com/blog/5"))

	if page, err := store.CurrentPage(); assert.Nil(err) {
		assert.Equal(page.Name, "page no 5")
		assert.Equal(page.URL, "http://example.com/blog/5")
	}

	if page, err := store.Page("page no 3"); assert.Nil(err) {
		assert.Equal(page.Name, "page no 3")
		assert.Equal(page.URL, "http://example.com/blog/3")
	}

	if page, err := store.PageBefore("page no 3"); assert.Nil(err) {
		assert.Equal(page.Name, "page no 2")
		assert.Equal(page.URL, "http://example.com/blog/2")
	}

	if page, err := store.PageAfter("page no 3"); assert.Nil(err) {
		assert.Equal(page.Name, "page no 4")
		assert.Equal(page.URL, "http://example.com/blog/4")
	}

	// list all
	if pages, err := store.Pages(); assert.Nil(err) {
		assert.Len(pages, 5)

		assert.Equal(pages[0].Name, "page no 5")
		assert.Equal(pages[4].Name, "what is this")
	}
}
