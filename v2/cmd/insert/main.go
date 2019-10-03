package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	// register sqlite3 for database/sql
	_ "github.com/mattn/go-sqlite3"

	"hawx.me/code/tally-ho/v2/blog"
)

func main() {
	dbPath := flag.String("db", "", "")
	flag.Parse()

	if *dbPath == "" {
		log.Println("--db PATH required")
		return
	}

	db, err := blog.Open(*dbPath)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	var data map[string][]interface{}
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		log.Println(err)
		return
	}

	if _, err := db.Create(data); err != nil {
		log.Println(err)
	}
}
