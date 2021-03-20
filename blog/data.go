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

	return b.withAuthor(groups[0].Properties), nil
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

	return b.withAuthor(groups[0].Properties), nil
}

func (b *Blog) Delete(url string) error {
	data, err := b.Entry(url)
	if err != nil {
		return err
	}

	id, ok := data["uid"][0].(string)
	if !ok {
		return errors.New("post to delete not found")
	}

	go b.sendWebmentions(url, data)
	go b.hubPublish()

	return b.entries.Set(id, "hx-deleted", true)
}

func (b *Blog) Undelete(url string) error {
	data, err := b.Entry(url)
	if err != nil {
		return err
	}

	id, ok := data["uid"][0].(string)
	if !ok {
		return errors.New("post to undelete not found")
	}

	go b.sendWebmentions(url, data)
	go b.hubPublish()

	return b.entries.DeletePredicate(id, "hx-deleted")
}

func (b *Blog) Mention(source string, data map[string][]interface{}) error {
	// TODO: add ability to block by host or url
	if err := b.mentions.DeleteSubject(source); err != nil {
		return err
	}

	if _, ok := data["hx-gone"]; ok {
		return nil
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

func (b *Blog) MentionsBefore(published time.Time, limit int) (list []numbersix.Group, err error) {
	triples, err := b.mentions.List(numbersix.
		Before("published", published.Format(time.RFC3339)).
		Limit(limit))
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

	return b.groupedWithAuthors(numbersix.Grouped(triples)), nil
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

	return b.groupedWithAuthors(numbersix.Grouped(triples)), nil
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

	return b.groupedWithAuthors(numbersix.Grouped(triples)), nil
}

func (b *Blog) LikesOn(ymd string) (groups []numbersix.Group, err error) {
	// TODO: this should be sorted
	triples, err := b.entries.List(
		numbersix.
			Begins("published", ymd).
			Without("hx-deleted").
			Has("like-of"),
	)
	if err != nil {
		return
	}

	return b.groupedWithAuthors(numbersix.Grouped(triples)), nil
}

func (b *Blog) withAuthor(m map[string][]interface{}) map[string][]interface{} {
	if _, ok := m["author"]; ok {
		return m
	}

	m["author"] = []interface{}{
		map[string]interface{}{
			"types": []interface{}{"h-card"},
			"properties": map[string][]interface{}{
				"name": {b.config.Name},
				"url":  {b.config.Me},
			},
		},
	}

	return m
}

func (b *Blog) groupedWithAuthors(gs []numbersix.Group) []numbersix.Group {
	for _, g := range gs {
		g.Properties = b.withAuthor(g.Properties)
	}

	return gs
}
