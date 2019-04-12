package blog

import (
	"errors"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"hawx.me/code/tally-ho/data"
)

type Blog struct {
	basePath, baseURL   string
	mediaPath, mediaURL string
	templates           *template.Template
	store               *data.Store
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

	// MediaURL is the URL that the media will be hosted at.
	MediaURL string

	// MediaPath is the path media files will be written to.
	MediaPath string
}

func New(options Options) (*Blog, error) {
	if len(options.BaseURL) == 0 {
		return nil, errors.New("BaseURL must be something")
	}
	if len(options.BaseURL) == 0 || options.BaseURL[len(options.BaseURL)-1] != '/' {
		return nil, errors.New("BaseURL must end with a '/'")
	}
	if len(options.BasePath) == 0 || options.BasePath[len(options.BasePath)-1] != '/' {
		return nil, errors.New("BasePath must end with a '/'")
	}
	if len(options.MediaURL) == 0 || options.MediaURL[len(options.MediaURL)-1] != '/' {
		return nil, errors.New("MediaURL must end with a '/'")
	}
	if len(options.MediaPath) == 0 || options.MediaPath[len(options.MediaPath)-1] != '/' {
		return nil, errors.New("MediaPath must end with a '/'")
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
		baseURL:   options.BaseURL,
		basePath:  options.BasePath,
		mediaURL:  options.MediaURL,
		mediaPath: options.MediaPath,
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
