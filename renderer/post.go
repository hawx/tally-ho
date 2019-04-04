package renderer

import (
	"database/sql"
	"html/template"
	"log"
	"os"
	"path/filepath"

	"hawx.me/code/tally-ho/config"
	"hawx.me/code/tally-ho/data"
)

func New(conf *config.Config, glob string) (*Renderer, error) {
	tmpls, err := template.New("ts").Funcs(template.FuncMap{
		"has": func(m map[string][]interface{}, key string) bool {
			value, ok := m[key]

			return ok && len(value) > 0
		},
		"getOr": func(m map[string][]interface{}, key string, or interface{}) interface{} {
			value, ok := m[key]

			if ok && len(value) > 0 {
				return value[0]
			}

			return or
		},
		"mustGet": func(m map[string][]interface{}, key string) interface{} {
			return m[key][0]
		},
		"content": func(m map[string][]interface{}) interface{} {
			contents, ok := m["content"]

			if ok && len(contents) > 0 {
				if content, ok := contents[0].(string); ok {
					return content
				}

				if content, ok := contents[0].(map[string]interface{}); ok {
					if html, ok := content["html"]; ok {
						return template.HTML(html.(string))
					}

					if text, ok := content["text"]; ok {
						return text
					}
				}
			}

			return ""
		},
	}).ParseGlob(glob)

	return &Renderer{conf: conf, tmpls: tmpls}, err
}

type Renderer struct {
	conf  *config.Config
	tmpls *template.Template
}

func (r *Renderer) All(store *data.Store) error {
	pages, err := store.Pages()
	if err != nil {
		return err
	}

	for i, page := range pages {
		entries, err := store.Entries(page.Name)
		if err != nil {
			return err
		}

		var posts []map[string][]interface{}
		for _, entry := range entries {
			properties := entry.Properties
			posts = append(posts, properties)
		}

		pageCtx := PageCtx{
			Title: page.Name,
			Posts: posts,
		}

		if i > 0 {
			pageCtx.NextPage = &pages[i-1]
		}
		if i < len(pages)-1 {
			pageCtx.PrevPage = &pages[i+1]
		}

		if err := r.simpleRenderPage(page.URL, pageCtx); err != nil {
			log.Println(err)
		}

		if i == 0 {
			if err := r.simpleRenderPage(r.conf.RootURL(), pageCtx); err != nil {
				log.Println(err)
			}
		}

		for _, post := range posts {
			if err := r.simpleRenderPost(post); err != nil {
				log.Println(err)
			}
		}
	}

	return nil
}

type PageCtx struct {
	Title    string
	Posts    []map[string][]interface{}
	NextPage *data.Page
	PrevPage *data.Page
}

// RenderPage will fully render the page.
func (r *Renderer) RenderPage(page data.Page, store *data.Store) error {
	entries, err := store.Entries(page.Name)
	if err != nil {
		return err
	}

	var posts []map[string][]interface{}
	for _, entry := range entries {
		properties := entry.Properties
		posts = append(posts, properties)
	}

	pageCtx := PageCtx{
		Title: page.Name,
		Posts: posts,
	}

	prevPage, err := store.PageBefore(page.Name)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		pageCtx.PrevPage = &prevPage
	}

	nextPage, err := store.PageAfter(page.Name)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		pageCtx.NextPage = &nextPage
	}

	return r.simpleRenderPage(page.URL, pageCtx)
}

// RenderPost is called when a new post is created. It is given the new post's
// id and renders the new post, the page the post is on, and a new
// /index.html. If this is the first post of the page, then also render the
// previous page so the next page link is created.
func (r *Renderer) RenderPost(id string, properties map[string][]interface{}, store *data.Store) error {
	log.Println("rendering", id)

	page, err := store.Page(properties["hx-page"][0].(string))
	if err != nil {
		return err
	}

	entries, err := store.Entries(page.Name)
	if err != nil {
		return err
	}

	var posts []map[string][]interface{}
	for _, entry := range entries {
		properties := entry.Properties
		posts = append(posts, properties)
	}

	pageCtx := PageCtx{
		Title: page.Name,
		Posts: posts,
	}

	prevPage, err := store.PageBefore(page.Name)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil {
		pageCtx.PrevPage = &prevPage
	}

	if err := r.simpleRenderPage(page.URL, pageCtx); err != nil {
		return err
	}

	if err := r.simpleRenderPage(r.conf.RootURL(), pageCtx); err != nil {
		return err
	}

	if len(posts) == 1 && prevPage.URL != "" {
		if err := r.RenderPage(prevPage, store); err != nil {
			return err
		}
	}

	err = r.simpleRenderPost(properties)
	return err
}

func (r *Renderer) simpleRenderPage(pageURL string, pageCtx PageCtx) error {
	path := r.conf.URLToPath(pageURL)

	return r.write(path, "page.gotmpl", pageCtx)
}

func (r *Renderer) simpleRenderPost(properties map[string][]interface{}) error {
	url := properties["url"][0].(string)
	path := r.conf.URLToPath(url)

	return r.write(path, "post.gotmpl", properties)
}

func (r *Renderer) write(path, tmpl string, v interface{}) error {
	dir := filepath.Dir(path)

	log.Println("mkdir", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	log.Println("writing", path)
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	return r.tmpls.ExecuteTemplate(file, tmpl, v)
}
