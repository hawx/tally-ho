package blog

import (
	"io"
	"os"

	"github.com/google/uuid"
)

func (b *Blog) WriteMedia(r io.Reader) (location string, err error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	file, err := os.Create(b.mediaPath + uid.String())
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, r); err != nil {
		return "", err
	}

	return b.mediaURL + uid.String(), nil
}
