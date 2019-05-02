package writer

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

func TestNewFileWriter(t *testing.T) {
	assert := assert.New(t)

	_, err := NewFileWriter("/tmp", "http://example.com/")
	assert.NotNil(err)

	_, err = NewFileWriter("/tmp/", "http://example.com")
	assert.NotNil(err)
}

func TestFileWriterCopyToFile(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "file-writer")
	if !assert.Nil(err) {
		return
	}
	defer os.RemoveAll(tmpDir)

	w, err := NewFileWriter(tmpDir+"/", "http://example.com/")
	assert.Nil(err)

	assert.Nil(w.CopyToFile("some/file.txt", strings.NewReader("hello there")))
	data, _ := ioutil.ReadFile(tmpDir + "/some/file.txt")
	assert.Equal("hello there", string(data))

	assert.Nil(w.CopyToFile("some/file.txt", strings.NewReader("hello there")))
	data, _ = ioutil.ReadFile(tmpDir + "/some/file.txt")
	assert.Equal("hello there", string(data))

	assert.NotNil(w.CopyToFile("/some/other/file.txt", strings.NewReader("the senate")))
}

func TestFileWriterURL(t *testing.T) {
	assert := assert.New(t)

	w, err := NewFileWriter("/tmp/", "http://example.com/")
	assert.Nil(err)

	assert.Equal("http://example.com/some/file.txt", w.URL("some/file.txt"))
	assert.Equal("http://example.com/some/", w.URL("some/index.html"))

	assert.Equal("http://example.com/some/file.txt", w.URL("/tmp/some/file.txt"))
	assert.Equal("http://example.com/some/", w.URL("/tmp/some/index.html"))

}

func TestFileWriterPath(t *testing.T) {
	assert := assert.New(t)

	w, err := NewFileWriter("/tmp/", "http://example.com/")
	assert.Nil(err)

	assert.Equal("/tmp/some/file.txt", w.Path("some/file.txt"))
	assert.Equal("/tmp/some/index.html", w.Path("some/"))

	assert.Equal("/tmp/some/file.txt", w.Path("http://example.com/some/file.txt"))
	assert.Equal("/tmp/some/index.html", w.Path("http://example.com/some/"))
}
