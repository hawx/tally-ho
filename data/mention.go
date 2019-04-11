package data

import (
	"net/url"

	"hawx.me/code/numbersix"
)

func (s *Store) AddMention(source string, data map[string][]interface{}) error {
	if err := s.mentions.DeleteSubject(source); err != nil {
		return err
	}

	return s.mentions.SetProperties(source, data)
}

func (s *Store) MentionSourceAllowed(source string) bool {
	any, err := s.mentions.Any(numbersix.About(source).Where("blocked", true))
	if err != nil || !any {
		return true
	}

	sourceURL, err := url.Parse(source)
	if err != nil {
		return false
	}

	any, err = s.mentions.Any(numbersix.About(sourceURL.Host).Where("blocked", true))
	if err != nil || !any {
		return true
	}

	return false
}
