package data

import (
	"errors"
	"time"

	"hawx.me/code/numbersix"
)

func (s *Store) Create(id string, data map[string][]interface{}) error {
	return s.entries.SetProperties(id, data)
}

func (s *Store) Update(id string, replace, add, delete map[string][]interface{}) error {
	replace["updated"] = []interface{}{time.Now().UTC().Format(time.RFC3339)}

	for predicate, values := range replace {
		s.entries.DeletePredicate(id, predicate)
		s.entries.SetMany(id, predicate, values)
	}

	for predicate, values := range add {
		s.entries.SetMany(id, predicate, values)
	}

	for predicate, values := range delete {
		for _, value := range values {
			s.entries.DeleteValue(id, predicate, value)
		}
	}

	return nil
}

func (s *Store) Get(id string) (data map[string][]interface{}, err error) {
	triples, err := s.entries.List(numbersix.About(id))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for id: " + id)
	}

	return groups[0].Properties, nil
}

func (s *Store) Entries(page string) (groups []numbersix.Group, err error) {
	triples, err := s.entries.List(numbersix.Descending("published").Where("hx-page", page))
	if err != nil {
		return
	}

	return numbersix.Grouped(triples), nil
}