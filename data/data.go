package data

import (
	"errors"

	"github.com/google/uuid"
	"hawx.me/code/numbersix"
)

func Open(path string) (*Store, error) {
	db, err := numbersix.Open(path)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

type Store struct {
	db *numbersix.DB
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Create(data map[string][]interface{}) (string, error) {
	id := uuid.New().String()

	return id, s.db.SetProperties(id, data)
}

func (s *Store) Update(id string, replace, add, delete map[string][]interface{}) error {
	for predicate, values := range replace {
		s.db.DeletePredicate(id, predicate)
		s.db.SetMany(id, predicate, values)
	}

	for predicate, values := range add {
		s.db.SetMany(id, predicate, values)
	}

	for predicate, values := range delete {
		for _, value := range values {
			s.db.DeleteValue(id, predicate, value)
		}
	}

	return nil
}

func (s *Store) Get(id string) (data map[string][]interface{}, err error) {
	triples, err := s.db.List(numbersix.About(id))
	if err != nil {
		return
	}
	groups := numbersix.Grouped(triples)
	if len(groups) == 0 {
		return data, errors.New("no data for id: " + id)
	}

	return groups[0].Properties, nil
}
