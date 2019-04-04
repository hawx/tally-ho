package config

import (
	"testing"

	"hawx.me/code/assert"
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)

	conf, err := New("http://example.com/blog/", "/wwwroot/blog/")
	assert.Nil(err)

	const (
		rootPath = "/wwwroot/blog/index.html"
		pageURL  = "http://example.com/blog/my-page"
		pagePath = "/wwwroot/blog/my-page/index.html"
		postURL  = "http://example.com/blog/my-page/my-post"
		postPath = "/wwwroot/blog/my-page/my-post/index.html"
	)

	assert.Equal("my-post", conf.PostID(postURL))
	assert.Equal(postURL, conf.PostURL(pageURL, "my-post"))
	assert.Equal(postPath, conf.URLToPath(postURL))
	assert.Equal(postURL, conf.PathToURL(postPath))
	assert.Equal(postPath, conf.PostPath("my-page", "my-post"))
	assert.Equal(pageURL, conf.PageURL("my-page"))
	assert.Equal(pagePath, conf.PagePath("my-page"))
	assert.Equal(rootPath, conf.URLToPath(conf.RootURL()))
}
