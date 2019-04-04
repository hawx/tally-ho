package data

type Page struct {
	Name string
	URL  string
}

func (s *Store) SetNextPage(name string) error {
	_, err := s.sqlite.Exec(`INSERT INTO pages (name, url) VALUES (?, ?)`,
		name,
		s.conf.PageURL(slugify(name)))

	return err
}

func (s *Store) CurrentPage() (page Page, err error) {
	row := s.sqlite.QueryRow(`SELECT name, url FROM pages ORDER BY id DESC LIMIT 1`)

	err = row.Scan(&page.Name, &page.URL)
	return
}

func (s *Store) Page(name string) (page Page, err error) {
	row := s.sqlite.QueryRow(`SELECT name, url FROM pages WHERE name = ?`, name)

	err = row.Scan(&page.Name, &page.URL)
	return
}

func (s *Store) PageBefore(name string) (page Page, err error) {
	row := s.sqlite.QueryRow(`
SELECT name, url FROM pages
WHERE id < (SELECT id FROM pages WHERE name = ?)
ORDER BY id DESC
LIMIT 1
`, name)

	err = row.Scan(&page.Name, &page.URL)
	return
}

func (s *Store) PageAfter(name string) (page Page, err error) {
	row := s.sqlite.QueryRow(`
SELECT name, url FROM pages
WHERE id > (SELECT id FROM pages WHERE name = ?)
ORDER BY id
LIMIT 1
`, name)

	err = row.Scan(&page.Name, &page.URL)
	return
}

func (s *Store) Pages() (pages []Page, err error) {
	rows, err := s.sqlite.Query(`SELECT name, url FROM pages ORDER BY id DESC`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var page Page
		if err = rows.Scan(&page.Name, &page.URL); err != nil {
			return
		}
		pages = append(pages, page)
	}

	return pages, rows.Err()
}
