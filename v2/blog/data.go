package blog

import (
	"database/sql"
	"errors"
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

	six, err := numbersix.For(db, "entries")
	if err != nil {
		return nil, err
	}

	return &DB{six: six}, nil
}

type DB struct {
	six *numbersix.DB
}

func (db *DB) Close() error {
	return db.six.Close()
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

	return data["url"][0].(string), db.six.SetProperties(id, data)
}

func (db *DB) Update(url string, replace, add, delete map[string][]interface{}) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	triples, err := db.six.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	id := triples[0].Subject

	for predicate, values := range replace {
		db.six.DeletePredicate(id, predicate)
		db.six.SetMany(id, predicate, values)
	}

	for predicate, values := range add {
		db.six.SetMany(id, predicate, values)
	}

	for predicate, values := range delete {
		for _, value := range values {
			db.six.DeleteValue(id, predicate, value)
		}
	}

	return nil
}

func (db *DB) Entry(url string) (data map[string][]interface{}, err error) {
	triples, err := db.six.List(numbersix.Where("url", url))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for url: " + url)
	}

	return groups[0].Properties, nil
}

func (db *DB) Before(published time.Time) (groups []numbersix.Group, err error) {
	triples, err := db.six.List(
		numbersix.
			Before("published", published.Format(time.RFC3339)).
			Limit(5),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}
