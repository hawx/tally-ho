package config

import (
	"errors"
	"net/url"
	"strings"
)

type Config struct {
	postsURL *url.URL
}

// New e.g. New("https://example.com/weblog/", "posts")
func New(base, postsPath string) (Config, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return Config{}, err
	}

	postsURL, err := baseURL.Parse(postsPath)

	return Config{postsURL: postsURL}, err
}

// ID takes a URL for a post and returns the ID.
func (c Config) ID(url string) (string, error) {
	if !strings.HasPrefix(url, c.postsURL.String()) {
		return "", errors.New("url does not match expected for posts")
	}

	id := url[len(c.postsURL.String()):]
	return id, nil
}

// Post takes an ID for a post and returns the URL.
func (c Config) Post(id string) (string, error) {
	postURL, err := c.postsURL.Parse(id)
	if err != nil {
		return "", err
	}

	return postURL.String(), nil
}
