package blog

import (
	"errors"
	"time"
)

func (b *Blog) Update(
	url string,
	replace, add, delete map[string][]interface{},
	deleteAll []string,
) error {
	oldData, err := b.Entry(url)
	if err != nil {
		return err
	}

	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	id, ok := oldData["uid"][0].(string)
	if !ok {
		return errors.New("post to update not found")
	}

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

	newData, err := b.Entry(url)
	if err != nil {
		return err
	}

	go b.sendUpdateWebmentions(url, oldData, newData)

	return nil
}
