package micropub

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/numbersix"
	"hawx.me/code/tally-ho/writer"
)

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS pages (
      id   INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT,
      url  TEXT
    );
  `)

	return err
}

type Page struct {
	Name, URL string
}

type Reader struct {
	db *micropubDB
}

func (r *Reader) Post(url string) (properties map[string][]interface{}, err error) {
	return r.db.entryByURL(url)
}

func (r *Reader) Entries(url string) (groups []numbersix.Group, err error) {
	page, err := r.Page(url)
	if err != nil {
		return
	}

	triples, err := r.db.entries.List(numbersix.Descending("published").Where("hx-page", page.Name))
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}

func (r *Reader) CurrentPage() (Page, error) {
	name, url, err := r.db.currentPage()

	return Page{Name: name, URL: url}, err
}

func (r *Reader) PageBefore(url string) (prev Page, err error) {
	row := r.db.sql.QueryRow(`
SELECT name, url FROM pages
WHERE id < (SELECT id FROM pages WHERE url = ?)
ORDER BY id DESC
LIMIT 1
`, url)

	err = row.Scan(&prev.Name, &prev.URL)
	return
}

func (r *Reader) PageAfter(url string) (next Page, err error) {
	row := r.db.sql.QueryRow(`
SELECT name, url FROM pages
WHERE id > (SELECT id FROM pages WHERE url = ?)
ORDER BY id
LIMIT 1
`, url)

	err = row.Scan(&next.Name, &next.URL)
	return
}

func (r *Reader) Page(url string) (page Page, err error) {
	row := r.db.sql.QueryRow(`SELECT name, url FROM pages WHERE url = ?`, url)

	err = row.Scan(&page.Name, &page.URL)
	return
}

type micropubDB struct {
	sql     *sql.DB
	entries *numbersix.DB
}

func (db *micropubDB) currentPage() (name, url string, err error) {
	row := db.sql.QueryRow(`SELECT name, url FROM pages ORDER BY id DESC LIMIT 1`)

	err = row.Scan(&name, &url)
	return
}

func (db *micropubDB) setNextPage(uf writer.URLFactory, name string) error {
	url := uf.URL(slugify(name) + "/")

	_, err := db.sql.Exec(`INSERT INTO pages (name, url) VALUES (?, ?)`,
		name,
		url)

	return err
}

func (db *micropubDB) createEntry(data map[string][]interface{}) (map[string][]interface{}, error) {
	id := uuid.New().String()

	pageName, pageURL, err := db.currentPage()
	if err != nil {
		return data, err
	}

	slug := id
	if len(data["name"]) == 1 {
		name := data["name"][0].(string)
		if len(name) > 0 {
			slug = slugify(name)
		}
	}
	if len(data["mp-slug"]) == 1 {
		slug = data["mp-slug"][0].(string)
	}

	data["uid"] = []interface{}{id}
	data["hx-page"] = []interface{}{pageName}
	data["url"] = []interface{}{postURL(pageURL, slug)}

	if len(data["published"]) == 0 {
		data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}
	}

	return data, db.entries.SetProperties(id, data)
}

func (db *micropubDB) updateEntry(url string, replace, add, delete map[string][]interface{}) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	triples, err := db.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	id := triples[0].Subject

	for predicate, values := range replace {
		db.entries.DeletePredicate(id, predicate)
		db.entries.SetMany(id, predicate, values)
	}

	for predicate, values := range add {
		db.entries.SetMany(id, predicate, values)
	}

	for predicate, values := range delete {
		for _, value := range values {
			db.entries.DeleteValue(id, predicate, value)
		}
	}

	return nil
}

func (db *micropubDB) entryByURL(url string) (data map[string][]interface{}, err error) {
	triples, err := db.entries.List(numbersix.Where("url", url))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for url: " + url)
	}

	return groups[0].Properties, nil
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

func postURL(pageURL, slug string) string {
	if pageURL[len(pageURL)-1] == '/' {
		return pageURL + slug + "/"
	}
	return pageURL + "/" + slug + "/"
}
