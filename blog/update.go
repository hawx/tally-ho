package blog

import (
	"errors"
	"time"
)

func (b *Blog) Update(
	url string,
	replace, add, delete map[string][]any,
	deleteAll []string,
) error {
	oldData, err := b.Entry(url)
	if err != nil {
		return err
	}

	replace["updated"] = []any{time.Now().UTC().Format(time.RFC3339)}

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

	b.massage(newData)

	if err := b.entries.DeleteSubject(id); err != nil {
		return err
	}
	if err := b.entries.SetProperties(id, newData); err != nil {
		return err
	}

	// don't send updates for published pages (although maybe I'll change this in the future)
	if len(newData["hx-url"]) == 0 {
		go b.sendUpdateWebmentions(url, oldData, newData)
		go b.hubPublish()
	}

	return nil
}
