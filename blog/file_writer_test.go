package blog

import (
	"testing"

	"hawx.me/code/assert"
)

func TestExtension(t *testing.T) {
	testCases := map[string]struct {
		ContentType, Filename, Expected string
	}{
		"from extension": {
			ContentType: "image/jpeg",
			Filename:    "file.extension",
			Expected:    ".extension",
		},
		"from uppercase extension": {
			ContentType: "image/jpeg",
			Filename:    "FILE.EXTENSION",
			Expected:    ".extension",
		},
		"from content-type": {
			ContentType: "image/jpeg",
			Filename:    "a-photo",
			Expected:    ".jpeg",
		},
		"from nothing": {
			ContentType: "",
			Filename:    "",
			Expected:    "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ext := extension(tc.ContentType, tc.Filename)
			assert.Equal(t, tc.Expected, ext)
		})
	}
}
