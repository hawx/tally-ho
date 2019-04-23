package blog

import (
	"errors"
	"html/template"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/tally-ho/data"
)

type Blog struct {
	fw        FileWriter
	baseURL   string
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
	fw, err := NewFileWriter(options.BasePath, options.BaseURL)
	if err != nil {
		return nil, err
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
		fw:        fw,
		baseURL:   options.BaseURL,
		templates: templates,
		store:     store,
	}, nil
}

func (b *Blog) Close() error {
	return b.store.Close()
}

// post.go

func (b *Blog) Update(url string, replace, add, delete map[string][]interface{}) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	return b.store.Update(url, replace, add, delete)
}

func (b *Blog) Rerender(url string) error {
	parts := strings.SplitAfter(url, "/")
	pageURL := strings.Join(parts[:len(parts)-2], "")

	page, err := FindPageByURL(b.baseURL, pageURL, b.store)
	if err != nil {
		return err
	}

	if err := page.Render(b.store, b.templates, b, false); err != nil {
		return err
	}

	for _, post := range page.Posts {
		if post["url"][0].(string) == url {
			post, _ := page.Post(post)
			return post.Render(b.templates, b)
		}
	}

	return errors.New("could not find post to render")
}

func (b *Blog) SetNextPage(name string) error {
	url := b.PageURL(slugify(name))

	return b.store.SetNextPage(name, url)
}

func (b *Blog) CurrentPage() (string, error) {
	page, err := b.store.CurrentPage()

	return page.Name, err
}

func (b *Blog) Create(data map[string][]interface{}) (map[string][]interface{}, error) {
	id := uuid.New().String()

	page, err := b.store.CurrentPage()
	if err != nil {
		return data, err
	}

	slug := id
	if len(data["name"]) == 1 {
		name := data["name"][0].(string)
		if len(name) > 0 {
			slug = slugify(name)
		}
	}
	if len(data["mp-slug"]) == 1 {
		slug = data["mp-slug"][0].(string)
	}

	data["uid"] = []interface{}{id}
	data["hx-page"] = []interface{}{page.Name}
	data["url"] = []interface{}{b.PostURL(page.URL, slug)}

	if len(data["published"]) == 0 {
		data["published"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}
	}

	return data, b.store.Create(id, data)
}

// configuration.go
func (b *Blog) PostByURL(url string) (map[string][]interface{}, error) {
	return b.store.GetByURL(url)
}

// mention.go

// MentionSourceAllowed will check if the source URL or host of the source URL
// has been blacklisted.
func (b *Blog) MentionSourceAllowed(source string) bool {
	return b.store.MentionSourceAllowed(source)
}

// AddMention will add the properties to a new webmention, or if a mention
// already exists for the source update those properties.
func (b *Blog) AddMention(source string, data map[string][]interface{}) error {
	return b.store.AddMention(source, data)
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

func (b *Blog) RenderAdmin(w io.Writer, data interface{}) error {
	return b.templates.ExecuteTemplate(w, "admin.gotmpl", data)
}
