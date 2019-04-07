package data

import (
	"database/sql"
	"errors"
	"time"

	"hawx.me/code/numbersix"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"
)

type urlFactory interface {
	PostURL(pageURL, slug string) string
	PageURL(pageSlug string) string
}

func Open(path string, conf urlFactory) (*Store, error) {
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db, err := numbersix.For(sqlite, "triples")
	if err != nil {
		return nil, err
	}

	return &Store{sqlite: sqlite, db: db, conf: conf}, migrate(sqlite)
}

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

type Store struct {
	sqlite *sql.DB
	db     *numbersix.DB
	conf   urlFactory
}

func (s *Store) Close() error {
	return s.sqlite.Close()
}

func (s *Store) Create(id string, data map[string][]interface{}) error {
	return s.db.SetProperties(id, data)
}

func (s *Store) Update(id string, replace, add, delete map[string][]interface{}) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	for predicate, values := range replace {
		s.db.DeletePredicate(id, predicate)
		s.db.SetMany(id, predicate, values)
	}

	for predicate, values := range add {
		s.db.SetMany(id, predicate, values)
	}

	for predicate, values := range delete {
		for _, value := range values {
			s.db.DeleteValue(id, predicate, value)
		}
	}

	return nil
}

func (s *Store) Get(id string) (data map[string][]interface{}, err error) {
	triples, err := s.db.List(numbersix.About(id))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for id: " + id)
	}

	return groups[0].Properties, nil
}

func (s *Store) Entries(page string) (groups []numbersix.Group, err error) {
	triples, err := s.db.List(numbersix.Descending("published").Where("hx-page", page))
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}
