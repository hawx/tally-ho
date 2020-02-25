// Package silos provides methods to re-publish entries on other sites.
package silos

import "fmt"

// ErrUnsure can be used when a syndicator is not able to determine what should
// be posted. It contains the data of the entry that was used.
type ErrUnsure struct {
	data map[string][]interface{}
}

func (e ErrUnsure) Error() string {
	return fmt.Sprintf("unsure what to create: %#v", e.data)
}
