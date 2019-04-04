package config

import (
	"errors"
	"strings"
)

const indexHTML = "index.html"

// Config contains URLs and filepaths for determining where things should live.
type Config struct {
	rootURL  string
	rootPath string
}

// New e.g. New("https://example.com/weblog/", "/wwwroot/weblog/")
func New(rootURL, rootPath string) (*Config, error) {
	if len(rootURL) == 0 {
		return nil, errors.New("rootURL must be something")
	}
	if rootURL[len(rootURL)-1] != '/' {
		return nil, errors.New("rootURL must end with a '/'")
	}
	if rootPath[len(rootPath)-1] != '/' {
		return nil, errors.New("rootPath must end with a '/'")
	}

	return &Config{
		rootURL:  rootURL,
		rootPath: rootPath,
	}, nil
}

// PostID takes a URL for a post and returns the ID.
func (c *Config) PostID(url string) string {
	parts := strings.Split(url, "/")

	return parts[len(parts)-1]
}

// PostURL takes an ID for a post and returns the URL.
func (c *Config) PostURL(pageURL, uid string) string {
	return pageURL + "/" + uid
}

func (c *Config) URLToPath(url string) string {
	if url == c.rootURL {
		return c.rootPath + indexHTML
	}

	return c.rootPath + url[len(c.rootURL):] + "/" + indexHTML
}

func (c *Config) PathToURL(path string) string {
	return c.rootURL + path[len(c.rootPath):len(path)-len(indexHTML)-1]
}

// PostPath takes an ID for a post and returns the path.
func (c *Config) PostPath(pageSlug, id string) string {
	return c.rootPath + pageSlug + "/" + id + "/" + indexHTML
}

func (c *Config) PageURL(pageSlug string) string {
	return c.rootURL + pageSlug
}

func (c *Config) PagePath(pageSlug string) string {
	return c.rootPath + pageSlug + "/" + indexHTML
}

func (c *Config) RootURL() string {
	return c.rootURL
}
