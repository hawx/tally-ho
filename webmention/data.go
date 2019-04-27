package webmention

import (
	"database/sql"
	"net/url"

	"hawx.me/code/numbersix"
)

type DB struct {
	mentions *numbersix.DB
}

func wrap(db *sql.DB) (*DB, error) {
	mentions, err := numbersix.For(db, "mentions")
	if err != nil {
		return nil, err
	}

	return &DB{mentions: mentions}, nil
}

func (db *DB) Upsert(source string, data map[string][]interface{}) error {
	if err := db.mentions.DeleteSubject(source); err != nil {
		return err
	}

	return db.mentions.SetProperties(source, data)
}

func (db *DB) AlloweddFromSource(source string) bool {
	any, err := db.mentions.Any(numbersix.About(source).Where("blocked", true))
	if err != nil || !any {
		return true
	}

	sourceURL, err := url.Parse(source)
	if err != nil {
		return false
	}

	any, err = db.mentions.Any(numbersix.About(sourceURL.Host).Where("blocked", true))
	if err != nil || !any {
		return true
	}

	return false
}
