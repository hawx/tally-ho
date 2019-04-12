package blog

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"hawx.me/code/assert"
)

func TestWriteMedia(t *testing.T) {
	assert := assert.New(t)

	re := regexp.MustCompile("^http://media.example.com/([a-f0-9\\-]{36})$")

	tmpDir, err := ioutil.TempDir("", "write-media")
	if !assert.Nil(err) {
		return
	}
	defer os.RemoveAll(tmpDir)

	blog := &Blog{
		mediaPath: tmpDir + "/",
		mediaURL:  "http://media.example.com/",
	}

	file := strings.NewReader("I am a file")

	location, err := blog.WriteMedia(file)
	assert.Nil(err)
	assert.Regexp(re, location)

	uid := re.FindAllStringSubmatch(location, -1)[0][1]
	data, err := ioutil.ReadFile(tmpDir + "/" + uid)
	assert.Nil(err)
	assert.Equal("I am a file", string(data))
}
