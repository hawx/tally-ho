package blog

import (
	"strings"
)

// PostID takes a URL for a post and returns the ID.
func (b *Blog) PostID(url string) string {
	parts := strings.Split(url, "/")
	if parts[len(parts)-1] == "" {
		return parts[len(parts)-2]
	}

	return parts[len(parts)-1]
}

// PostURL takes an ID for a post and returns the URL.
func (b *Blog) PostURL(pageURL, uid string) string {
	if pageURL[len(pageURL)-1] == '/' {
		return pageURL + uid + "/"
	}
	return pageURL + "/" + uid + "/"
}

func (b *Blog) URLToPath(url string) string {
	return b.fw.Path(url)
}

func (b *Blog) PathToURL(path string) string {
	return b.fw.URL(path)
}

// PostPath takes an ID for a post and returns the path.
func (b *Blog) PostPath(pageSlug, id string) string {
	return b.fw.Path(pageSlug + "/" + id + "/")
}

func (b *Blog) PageURL(pageSlug string) string {
	return b.fw.URL(pageSlug + "/")
}

func (b *Blog) PagePath(pageSlug string) string {
	return b.fw.Path(pageSlug + "/")
}

func (b *Blog) RootURL() string {
	return b.fw.URL("/")
}
