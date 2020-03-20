package blog

import (
	"io"
	"log"
	"mime"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
)

func (b *Blog) WriteFile(name, contentType string, r io.Reader) (location string, err error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	name = uid.String() + extension(contentType, name)
	p := path.Join(b.config.MediaDir, name)

	file, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, r); err != nil {
		return "", err
	}
	log.Printf("INFO wrote-file path=%s\n", p)

	relURL, _ := url.Parse(name)
	return b.config.MediaURL.ResolveReference(relURL).String(), nil
}

func extension(contentType, filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	if len(ext) > 0 {
		return ext
	}

	exts, err := mime.ExtensionsByType(contentType)
	if err == nil && len(exts) > 0 {
		return exts[0]
	}

	return ""
}
