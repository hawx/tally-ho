package blog

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
)

func (page *Page) Post(properties map[string][]interface{}) (*Post, error) {
	return &Post{
		Properties: properties,
		PageURL:    page.URL,
	}, nil
}

type Post struct {
	// Properties of the post.
	Properties map[string][]interface{}

	// PageURL is the full URL of the page the post belongs to.
	PageURL string
}

func (p *Post) Render(tmpl *template.Template, conf *Blog) error {
	url := p.Properties["url"][0].(string)
	path := conf.URLToPath(url)
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

	return tmpl.ExecuteTemplate(file, "post.gotmpl", p)
}
