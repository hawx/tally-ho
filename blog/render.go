package blog

import (
	"html/template"
	"log"

	"hawx.me/code/tally-ho/data"
)

// RenderPost can be called when a new post is added. It will render the post,
// the page the post is on, and the root page.
func RenderPost(
	properties map[string][]interface{},
	store *data.Store,
	tmpl *template.Template,
	conf *Config,
) error {
	log.Println("rendering", properties["uid"][0].(string))

	page, err := FindPage(properties["hx-page"][0].(string), store)
	if err != nil {
		return err
	}

	if err := page.Render(store, tmpl, conf, true); err != nil {
		return err
	}

	post, err := page.Post(properties)
	if err != nil {
		return err
	}

	if err := post.Render(tmpl, conf); err != nil {
		return err
	}

	return nil
}

// RenderAll will render the whole site, possibly again, overwriting any old
// files.
func RenderAll(store *data.Store, tmpl *template.Template, conf *Config) error {
	pages, err := store.Pages()
	if err != nil {
		return err
	}

	for _, page := range pages {
		page, err := FindPage(page.Name, store)
		if err != nil {
			return err
		}

		if err := page.Render(store, tmpl, conf, false); err != nil {
			return err
		}

		for _, properties := range page.Posts {
			post, err := page.Post(properties)
			if err != nil {
				return err
			}

			if err := post.Render(tmpl, conf); err != nil {
				return err
			}
		}
	}

	return nil
}
