package blog

import (
	"errors"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/tally-ho/data"
)

type Blog struct {
	baseURL   string
	basePath  string
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
	if len(options.BaseURL) == 0 {
		return nil, errors.New("BaseURL must be something")
	}
	if options.BaseURL[len(options.BaseURL)-1] != '/' {
		return nil, errors.New("BaseURL must end with a '/'")
	}
	if options.BasePath[len(options.BasePath)-1] != '/' {
		return nil, errors.New("BasePath must end with a '/'")
	}

	templates, err := parseTemplates(filepath.Join(options.WebPath, "template/*.gotmpl"))
	if err != nil {
		return nil, err
	}

	store, err := data.Open(options.DbPath, nil) // TODO: fix this nil
	if err != nil {
		return nil, err
	}

	return &Blog{
		basePath:  options.BasePath,
		baseURL:   options.BaseURL,
		templates: templates,
		store:     store,
	}, nil
}

func (b *Blog) Close() error {
	return b.store.Close()
}

// post.go

func (b *Blog) Update(id string, replace, add, delete map[string][]interface{}) error {
	return b.store.Update(id, replace, add, delete)
}

func (b *Blog) SetNextPage(name string) error {
	url := b.PageURL(slugify(name))

	return b.store.SetNextPage(name, url)
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
	data["url"] = []interface{}{b.PostURL(page.URL, slug)}
	data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	return data, b.store.Create(id, data)
}

// configuration.go
func (b *Blog) PostByURL(url string) (map[string][]interface{}, error) {
	id := b.PostID(url)

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

type writer interface {
	writePost(url string, data interface{}) error
	writePage(url string, data interface{}) error
	writeRoot(data interface{}) error
}

func (b *Blog) writePost(url string, data interface{}) error {
	return b.write(url, "post.gotmpl", data)
}

func (b *Blog) writePage(url string, data interface{}) error {
	return b.write(url, "page.gotmpl", data)
}

func (b *Blog) writeRoot(data interface{}) error {
	return b.write(b.RootURL(), "page.gotmpl", data)
}

func (b *Blog) write(url, tmpl string, data interface{}) error {
	path := b.URLToPath(url)
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

	return b.templates.ExecuteTemplate(file, tmpl, data)
}
