package syndicate

import "fmt"

type ErrUnsure struct {
	data map[string][]interface{}
}

func (e ErrUnsure) Error() string {
	return fmt.Sprintf("unsure what to create: %#v", e.data)
}

type Syndicator interface {
	Create(data map[string][]interface{}) (location string, err error)
	UID() string
	Name() string
}
