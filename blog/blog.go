package blog

import (
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/tally-ho/data"
)

type Blog struct {
	config    *Config
	templates *template.Template
	store     *data.Store
}

type Options struct {
	// WebPath is the path to the 'web' directory.
	WebPath string

	// BaseURL is the URL that the blog will be hosted at.
	BaseURL string

	// BasePath is the path the site will be written to.
	BasePath string

	// DbPath is the path to the sqlite database.
	DbPath string
}

func New(options Options) (*Blog, error) {
	config, err := NewConfig(options.BaseURL, options.BasePath)
	if err != nil {
		return nil, err
	}

	templates, err := ParseTemplates(filepath.Join(options.WebPath, "template/*.gotmpl"))
	if err != nil {
		return nil, err
	}

	store, err := data.Open(options.DbPath, config)
	if err != nil {
		return nil, err
	}

	return &Blog{
		config:    config,
		templates: templates,
		store:     store,
	}, nil
}

func (b *Blog) Close() error {
	return b.store.Close()
}

// post.go
func (b *Blog) PostID(url string) string {
	return b.config.PostID(url)
}

func (b *Blog) Update(id string, replace, add, delete map[string][]interface{}) error {
	return b.store.Update(id, replace, add, delete)
}

func (b *Blog) SetNextPage(name string) error {
	return b.store.SetNextPage(name)
}

func (b *Blog) Create(data map[string][]interface{}) (map[string][]interface{}, error) {
	id := uuid.New().String()

	page, err := b.store.CurrentPage()
	if err != nil {
		return data, err
	}

	slug := id
	if len(data["name"]) == 1 {
		slug = slugify(data["name"][0].(string))
	}
	if len(data["mp-slug"]) == 1 {
		slug = data["mp-slug"][0].(string)
	}

	data["uid"] = []interface{}{id}
	data["hx-page"] = []interface{}{page.Name}
	data["url"] = []interface{}{b.config.PostURL(page.URL, slug)}
	data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	return data, b.store.Create(id, data)
}

// configuration.go
func (b *Blog) PostByURL(url string) (map[string][]interface{}, error) {
	id := b.config.PostID(url)

	return b.store.Get(id)
}

var nonWord = regexp.MustCompile("\\W+")

func slugify(s string) string {
	s = strings.ReplaceAll(s, "'", "")
	s = nonWord.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	return s
}
