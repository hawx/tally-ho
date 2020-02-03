package blog

import (
	"errors"
	"time"

	"hawx.me/code/numbersix"
)

var empty = map[string][]interface{}{}

func (b *Blog) Entry(url string) (data map[string][]interface{}, err error) {
	triples, err := b.entries.List(numbersix.Where("url", url))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for url: " + url)
	}

	return groups[0].Properties, nil
}

func (b *Blog) EntryByUID(uid string) (data map[string][]interface{}, err error) {
	triples, err := b.entries.List(numbersix.Where("uid", uid))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for uid: " + uid)
	}

	return groups[0].Properties, nil
}

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

func (b *Blog) Delete(url string) error {
	triples, err := b.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	if len(triples) == 0 {
		return errors.New("post to delete not found")
	}
	id := triples[0].Subject

	return b.entries.Set(id, "hx-deleted", true)
}

func (b *Blog) Undelete(url string) error {
	triples, err := b.entries.List(numbersix.Where("url", url))
	if err != nil {
		return err
	}
	if len(triples) == 0 {
		return errors.New("post to undelete not found")
	}
	id := triples[0].Subject

	return b.entries.DeletePredicate(id, "hx-deleted")
}

func (b *Blog) Mention(source string, data map[string][]interface{}) error {
	// TODO: add ability to block by host or url
	if err := b.mentions.DeleteSubject(source); err != nil {
		return err
	}

	return b.mentions.SetProperties(source, data)
}

func (b *Blog) MentionsForEntry(url string) (list []numbersix.Group, err error) {
	triples, err := b.mentions.List(numbersix.Where("hx-target", url))
	if err != nil {
		return
	}

	list = numbersix.Grouped(triples)
	return
}

func (b *Blog) Before(published time.Time) (groups []numbersix.Group, err error) {
	triples, err := b.entries.List(
		numbersix.
			Before("published", published.Format(time.RFC3339)).
			Without("hx-deleted").
			Limit(25),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}

func (b *Blog) KindBefore(kind string, published time.Time) (groups []numbersix.Group, err error) {
	triples, err := b.entries.List(
		numbersix.
			Before("published", published.Format(time.RFC3339)).
			Where("hx-kind", kind).
			Without("hx-deleted").
			Limit(25),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}

func (b *Blog) CategoryBefore(category string, published time.Time) (groups []numbersix.Group, err error) {
	triples, err := b.entries.List(
		numbersix.
			Before("published", published.Format(time.RFC3339)).
			Where("category", category).
			Without("hx-deleted").
			Limit(25),
	)
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}

func (b *Blog) LikesOn(ymd string) (groups []numbersix.Group, err error) {
	triples, err := b.entries.List(
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
