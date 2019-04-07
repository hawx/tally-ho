package blog

import (
	"strings"
)

const indexHTML = "index.html"

// PostID takes a URL for a post and returns the ID.
func (b *Blog) PostID(url string) string {
	parts := strings.Split(url, "/")

	return parts[len(parts)-1]
}

// PostURL takes an ID for a post and returns the URL.
func (b *Blog) PostURL(pageURL, uid string) string {
	return pageURL + "/" + uid
}

func (b *Blog) URLToPath(url string) string {
	if url == b.baseURL {
		return b.basePath + indexHTML
	}

	return b.basePath + url[len(b.baseURL):] + "/" + indexHTML
}

func (b *Blog) PathToURL(path string) string {
	return b.baseURL + path[len(b.basePath):len(path)-len(indexHTML)-1]
}

// PostPath takes an ID for a post and returns the path.
func (b *Blog) PostPath(pageSlug, id string) string {
	return b.basePath + pageSlug + "/" + id + "/" + indexHTML
}

func (b *Blog) PageURL(pageSlug string) string {
	return b.baseURL + pageSlug
}

func (b *Blog) PagePath(pageSlug string) string {
	return b.basePath + pageSlug + "/" + indexHTML
}

func (b *Blog) RootURL() string {
	return b.baseURL
}
