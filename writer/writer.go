package writer

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileWriter is my new dream of what everything here can be built around. Given
// a path it can tell you how it is addressed on the web by URL, or it can write
// to a file somewhere. Also this should be easily mockable.
type FileWriter interface {
	CopyToFile(path string, r io.Reader) error
	URLFactory
	PathFactory
}

type URLFactory interface {
	URL(path string) string
}

type PathFactory interface {
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
