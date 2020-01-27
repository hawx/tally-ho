package blog

import (
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/numbersix"
)

var empty = map[string][]interface{}{}

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	entries, err := numbersix.For(db, "entries")
	if err != nil {
		return nil, err
	}

	mentions, err := numbersix.For(db, "mentions")
	if err != nil {
		return nil, err
	}

	return &DB{
		closer:   db,
		entries:  entries,
		mentions: mentions,
	}, nil
}

type DB struct {
	closer   io.Closer
	entries  *numbersix.DB
	mentions *numbersix.DB
}

func (db *DB) Close() error {
	return db.closer.Close()
}

func (db *DB) Create(data map[string][]interface{}) (location string, err error) {
	id := uuid.New().String()

	data["uid"] = []interface{}{id}
	// Use /entry/UID as canonical url so I don't have to change it in the future
	// if I decide on a nicer scheme. This will always be accessible and other
	// things can be layers on top of this.
	data["url"] = []interface{}{"/entry/" + id}

	if len(data["mp-slug"]) == 0 && len(data["name"]) > 0 {
		name := data["name"][0].(string)
		if len(name) > 0 {
			data["mp-slug"] = []interface{}{slugify(name)}
		}
	}

	if len(data["published"]) == 0 {
		data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}
	}

	return data["url"][0].(string), db.entries.SetProperties(id, data)
}

func (db *DB) Update(
	url string,
	replace, add, delete map[string][]interface{},
	deleteAll []string,
) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	triples, err := db.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	if len(triples) == 0 {
		return errors.New("post to update not found")
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

	for _, predicate := range deleteAll {
		db.entries.DeletePredicate(id, predicate)
	}

	return nil
}

func (db *DB) Delete(url string) error {
	triples, err := db.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	if len(triples) == 0 {
		return errors.New("post to delete not found")
	}
	id := triples[0].Subject

	return db.entries.Set(id, "hx-deleted", true)
}

func (db *DB) Undelete(url string) error {
	triples, err := db.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	if len(triples) == 0 {
		return errors.New("post to undelete not found")
	}
	id := triples[0].Subject

	return db.entries.DeletePredicate(id, "hx-deleted")
}

func (db *DB) Mention(source string, data map[string][]interface{}) error {
	// TODO: add ability to block by host or url
	if err := db.mentions.DeleteSubject(source); err != nil {
		return err
	}

	return db.mentions.SetProperties(source, data)
}

func (db *DB) Entry(url string) (data map[string][]interface{}, err error) {
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

func (db *DB) MentionsForEntry(url string) (list []numbersix.Group, err error) {
	triples, err := db.mentions.List(numbersix.Where("hx-target", url))
	if err != nil {
		return
	}

	list = numbersix.Grouped(triples)
	return
}

func (db *DB) Before(published time.Time) (groups []numbersix.Group, err error) {
	triples, err := db.entries.List(
		numbersix.
			Before("published", published.Format(time.RFC3339)).
			Without("hx-deleted").
			Limit(20),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}

func (db *DB) LikesOn(ymd string) (groups []numbersix.Group, err error) {
	triples, err := db.entries.List(
		numbersix.
			Begins("published", ymd).
			Without("hx-deleted").
			Has("like-of"),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}
