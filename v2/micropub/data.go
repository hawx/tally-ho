package micropub

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/numbersix"
)

type Reader struct {
	db *micropubDB
}

func (r *Reader) Post(id string) (properties map[string][]interface{}, err error) {
	return r.db.entryByID(id)
}

func (r *Reader) Before(published time.Time) (groups []numbersix.Group, err error) {
	triples, err := r.db.entries.List(
		numbersix.
			Before("published", published.Format(time.RFC3339)).
			Limit(5),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}

type micropubDB struct {
	sql     *sql.DB
	entries *numbersix.DB
}

func (db *micropubDB) createEntry(data map[string][]interface{}) (map[string][]interface{}, error) {
	id := uuid.New().String()

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
	data["url"] = []interface{}{slug} // TODO: check that the slug doesn't exist

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

func (db *micropubDB) entryByID(id string) (data map[string][]interface{}, err error) {
	triples, err := db.entries.List(numbersix.About(id))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for id: " + id)
	}

	return groups[0].Properties, nil
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
