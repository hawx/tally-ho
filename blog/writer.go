package blog

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FileWriter is my new dream of what everything here can be built around. Given
// a path it can tell you how it is addressed on the web by URL, or it can write
// to a file somewhere. Also this should be easily mockable.
type FileWriter interface {
	CopyToFile(path string, r io.Reader) error
	URL(path string) string
	Path(url string) string
}

type aFileWriter struct {
	dir string
	url string
}

// NewFileWriter returns a FileWriter that will create files in the directory,
// and construct URLs rooted at the url.
func NewFileWriter(dir, url string) (FileWriter, error) {
	if len(dir) == 0 || dir[len(dir)-1] != '/' {
		return nil, errors.New("dir must end with a '/'")
	}
	if len(url) == 0 || url[len(url)-1] != '/' {
		return nil, errors.New("url must end with a '/'")
	}

	return aFileWriter{dir: dir, url: url}, nil
}

func (w aFileWriter) CopyToFile(path string, r io.Reader) error {
	if path[0] == '/' {
		return errors.New("path to copy to must be relative")
	}

	dir := filepath.Dir(w.dir + path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(w.dir + path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, r)
	return err
}

func (w aFileWriter) URL(path string) string {
	if strings.HasPrefix(path, w.dir) {
		path = path[len(w.dir):]
	}

	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	return w.url + strings.TrimSuffix(path, "index.html")
}

func (w aFileWriter) Path(url string) string {
	if strings.HasPrefix(url, w.url) {
		url = url[len(w.url):]
	}

	if len(url) == 0 || url[len(url)-1] == '/' {
		url += "index.html"
	}

	return w.dir + url
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
