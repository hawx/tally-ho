package blog

import (
	"net/url"
	"testing"
	"time"

	"hawx.me/code/assert"
)

type Assert func(actual interface{}) *assert.WrappedAssertions

func TestMassage(t *testing.T) {
	baseURL, _ := url.Parse("http://example.com/")

	b := &Blog{
		config: Config{
			BaseURL: baseURL,
		},
	}

	testCases := map[string]struct {
		in map[string][]interface{}
		fn func(Assert, map[string][]interface{})
	}{
		"empty": {
			in: map[string][]interface{}{},
			fn: func(assert Assert, data map[string][]interface{}) {
				assert(data["uid"][0].(string)).NotEmpty()
				assert(data["url"][0].(string)).Equal("http://example.com/entry/" + data["uid"][0].(string))

				published, _ := time.Parse(time.RFC3339, data["published"][0].(string))
				assert(published).WithinDuration(time.Now(), time.Second)

				assert(data["hx-kind"][0].(string)).Equal("note")
			},
		},
		"non-Z-published": {
			in: map[string][]interface{}{
				"published": {"2020-10-01T12:03:01+0000"},
			},
			fn: func(assert Assert, data map[string][]interface{}) {
				published, _ := time.Parse(time.RFC3339, data["published"][0].(string))
				assert(published).Equal(time.Date(2020, time.October, 1, 12, 03, 1, 0, time.UTC))
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			b.massage(tc.in)
			tc.fn(assert.Wrap(t), tc.in)
		})
	}
}
