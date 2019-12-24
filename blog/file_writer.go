package blog

import (
	"io"
	"log"
	"net/url"
	"os"
	"path"
)

func (b *Blog) WriteFile(name string, r io.Reader) (location string, err error) {
	p := path.Join(b.MediaDir, name)

	file, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, r); err != nil {
		return "", err
	}
	log.Printf("INFO wrote-file path=%s\n", p)

	relURL, _ := url.Parse(path.Join("-", "media", name))
	return b.Config.BaseURL.ResolveReference(relURL).String(), nil
}
