package blog

import (
	"log"
	"os"
	"path/filepath"
)

type writer2 interface {
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
	return b.write(b.FileWriter.URL("/"), "page.gotmpl", data)
}

func (b *Blog) write(url, tmpl string, data interface{}) error {
	path := b.FileWriter.Path(url)
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

	return b.Templates.ExecuteTemplate(file, tmpl, data)
}
