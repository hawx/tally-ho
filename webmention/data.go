package webmention

import (
	"log"
	"net/url"

	"hawx.me/code/numbersix"
)

type Reader struct {
	db *numbersix.DB
}

// ForPost returns all webmentions that references a post.
func (r Reader) ForPost(url string) (list []numbersix.Group, err error) {
	triples, err := r.db.List(numbersix.Where("hx-target", url))
	if err != nil {
		return
	}

	list = numbersix.Grouped(triples)
	return
}

func upsertMention(db *numbersix.DB, source string, data map[string][]interface{}) error {
	if err := db.DeleteSubject(source); err != nil {
		return err
	}

	log.Println("upserting")
	return db.SetProperties(source, data)
}

func allowedFromSource(db *numbersix.DB, source string) bool {
	any, err := db.Any(numbersix.About(source).Where("blocked", true))
	if err != nil || !any {
		return true
	}

	sourceURL, err := url.Parse(source)
	if err != nil {
		return false
	}

	any, err = db.Any(numbersix.About(sourceURL.Host).Where("blocked", true))
	if err != nil || !any {
		return true
	}

	return false
}
