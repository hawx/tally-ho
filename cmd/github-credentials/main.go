package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/BurntSushi/toml"
	"golang.org/x/oauth2"
)

func main() {
	var conf struct {
		Github struct {
			ClientID     string
			ClientSecret string
		}
	}

	configPath := flag.String("config", "./config.toml", "")
	flag.Parse()

	if _, err := toml.DecodeFile(*configPath, &conf); err != nil {
		log.Println("ERR decode-config;", err)
		return
	}

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     conf.Github.ClientID,
		ClientSecret: conf.Github.ClientSecret,
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}

	next := make(chan struct{}, 0)

	srv := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			code := r.FormValue("code")

			tok, err := config.Exchange(ctx, code)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("accessToken =", tok.AccessToken)

			next <- struct{}{}
		}),
	}

	go srv.ListenAndServe()

	url := config.AuthCodeURL("some-random-state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	<-next
	srv.Shutdown(ctx)
}
