package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	token := flag.String("token", "", "")
	flag.Parse()

	if *token == "" {
		log.Println("--token TOKEN required")
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:8080/-/micropub", os.Stdin)
	if err != nil {
		log.Println(err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+*token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(resp)
}
