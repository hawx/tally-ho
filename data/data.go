// Package data wraps the underlying database used by the blog.
//
// It uses three tables.
//
// 1. pages
//
// This is a simple list of page names and urls.
//
// 2. entries
//
// This contains the blog posts and stuff. It stores everything as triples
// assuming that the input is micropubbed in the x-url-encoded format.
//
// 3. mentions
//
// Contains webmentions also stored as triples from the parsed h-entry that
// mentioned a entry.
package data

import (
	"database/sql"

	"hawx.me/code/numbersix"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"
)

type urlFactory interface {
	PostURL(pageURL, slug string) string
}

func Open(path string, conf urlFactory) (*Store, error) {
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	entries, err := numbersix.For(sqlite, "entries")
	if err != nil {
		return nil, err
	}

	mentions, err := numbersix.For(sqlite, "mentions")
	if err != nil {
		return nil, err
	}

	return &Store{
		sqlite:   sqlite,
		entries:  entries,
		mentions: mentions,
		conf:     conf,
	}, migrate(sqlite)
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
	sqlite   *sql.DB
	entries  *numbersix.DB
	mentions *numbersix.DB
	conf     urlFactory
}

func (s *Store) Close() error {
	return s.sqlite.Close()
}
