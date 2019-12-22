package syndicate

import "errors"

var ErrUnsure = errors.New("unsure what to create")

type Syndicator interface {
	Create(data map[string][]interface{}) (location string, err error)
	UID() string
	Name() string
}
