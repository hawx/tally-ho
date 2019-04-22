package blog

import (
	"log"
)

// RenderPost can be called when a new post is added. It will render the post,
// the page the post is on, and the root page.
func (b *Blog) RenderPost(properties map[string][]interface{}) error {
	log.Println("rendering", properties["uid"][0].(string))

	page, err := FindPage(b.baseURL, properties["hx-page"][0].(string), b.store)
	if err != nil {
		return err
	}

	if err := page.Render(b.store, b.templates, b, true); err != nil {
		return err
	}

	post, err := page.Post(properties)
	if err != nil {
		return err
	}

	if err := post.Render(b.templates, b); err != nil {
		return err
	}

	return nil
}

// RenderAll will render the whole site, possibly again, overwriting any old
// files.
func (b *Blog) RenderAll() error {
	pages, err := b.store.Pages()
	if err != nil {
		return err
	}

	for _, page := range pages {
		page, err := FindPage(b.baseURL, page.Name, b.store)
		if err != nil {
			return err
		}

		if err := page.Render(b.store, b.templates, b, false); err != nil {
			return err
		}

		for _, properties := range page.Posts {
			post, err := page.Post(properties)
			if err != nil {
				return err
			}

			if err := post.Render(b.templates, b); err != nil {
				return err
			}
		}
	}

	return nil
}
