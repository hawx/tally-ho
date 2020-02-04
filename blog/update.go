package blog

import (
	"errors"
	"time"

	"hawx.me/code/numbersix"
)

func (b *Blog) Update(
	url string,
	replace, add, delete map[string][]interface{},
	deleteAll []string,
) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	triples, err := b.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	if len(triples) == 0 {
		return errors.New("post to update not found")
	}
	id := triples[0].Subject

	for predicate, values := range replace {
		b.entries.DeletePredicate(id, predicate)
		b.entries.SetMany(id, predicate, values)
	}

	for predicate, values := range add {
		b.entries.SetMany(id, predicate, values)
	}

	for predicate, values := range delete {
		for _, value := range values {
			b.entries.DeleteValue(id, predicate, value)
		}
	}

	for _, predicate := range deleteAll {
		b.entries.DeletePredicate(id, predicate)
	}

	return nil
}
