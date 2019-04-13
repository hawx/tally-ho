package blog

import (
	"testing"

	"hawx.me/code/assert"
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)

	blog := &Blog{
		fw: aFileWriter{
			url: "http://example.com/blog/",
			dir: "/wwwroot/blog/",
		},
	}

	const (
		rootPath = "/wwwroot/blog/index.html"
		pageURL  = "http://example.com/blog/my-page/"
		pagePath = "/wwwroot/blog/my-page/index.html"
		postURL  = "http://example.com/blog/my-page/my-post/"
		postPath = "/wwwroot/blog/my-page/my-post/index.html"
	)

	assert.Equal("my-post", blog.PostID(postURL))
	assert.Equal(postURL, blog.PostURL(pageURL, "my-post"))
	assert.Equal(postPath, blog.URLToPath(postURL))
	assert.Equal(postURL, blog.PathToURL(postPath))
	assert.Equal(postPath, blog.PostPath("my-page", "my-post"))
	assert.Equal(pageURL, blog.PageURL("my-page"))
	assert.Equal(pagePath, blog.PagePath("my-page"))
	assert.Equal(rootPath, blog.URLToPath(blog.RootURL()))
}
