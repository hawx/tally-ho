package config

import (
	"errors"
	"net/url"
	"path/filepath"
	"strings"
)

// Config contains URLs and filepaths for determining where things should live.
type Config struct {
	postsURL  *url.URL
	postsPath string
}

// New e.g. New("https://example.com/weblog/", "/wwwroot/weblog", "posts")
func New(rootURL, rootPath, postsPath string) (Config, error) {
	baseURL, err := url.Parse(rootURL)
	if err != nil {
		return Config{}, err
	}

	postsURL, err := baseURL.Parse(postsPath)

	return Config{
		postsURL:  postsURL,
		postsPath: filepath.Join(rootPath, postsPath),
	}, err
}

// PostID takes a URL for a post and returns the ID.
func (c Config) PostID(url string) (string, error) {
	if !strings.HasPrefix(url, c.postsURL.String()) {
		return "", errors.New("url does not match expected for posts")
	}

	id := url[len(c.postsURL.String()):]
	return id, nil
}

// PostURL takes an ID for a post and returns the URL.
func (c Config) PostURL(id string) (string, error) {
	postURL, err := c.postsURL.Parse(id)
	if err != nil {
		return "", err
	}

	return postURL.String(), nil
}

// PostPath takes an ID for a post and returns the URL.
func (c Config) PostPath(id string) string {
	return filepath.Join(c.postsPath, id)
}
