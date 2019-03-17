package main

import (
	"flag"
	"log"
	"net/url"

	"github.com/google/uuid"
	"hawx.me/code/mux"
	"hawx.me/code/numbersix"
	"hawx.me/code/route"
	"hawx.me/code/serve"
	"hawx.me/code/tally-ho/handler"
)

func main() {
	var (
		port   = flag.String("port", "8080", "")
		socket = flag.String("socket", "", "")
		me     = flag.String("me", "", "")
		dbPath = flag.String("db", "file::memory:", "")
		base   = flag.String("base-url", "http://localhost:8080/", "")
	)
	flag.Parse()

	if *me == "" {
		log.Fatal("--me must be provided")
	}

	store, err := newStore(*dbPath)
	if err != nil {
		log.Fatal(err)
	}

	baseURL, err := url.Parse(*base)
	if err != nil {
		log.Fatal(err)
	}

	route.Handle("/micropub", mux.Method{
		"POST": handler.Authenticate(*me, "create", handler.Post(store, baseURL)),
	})

	route.Handle("/webmention", mux.Method{
		// "POST":
	})

	serve.Serve(*port, *socket, route.Default)
}

func newStore(path string) (*Store, error) {
	db, err := numbersix.Open(path)
	if err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

type Store struct {
	db *numbersix.DB
}

func (s *Store) Create(data map[string][]interface{}) (string, error) {
	id := uuid.New().String()

	return id, s.db.SetProperties(id, data)
}

func (s *Store) Update(id string, replace, add, delete map[string][]interface{}) error {
	return nil
}
