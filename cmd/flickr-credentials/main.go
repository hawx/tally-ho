package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/BurntSushi/toml"
	"github.com/gomodule/oauth1/oauth"
)

func main() {
	var conf struct {
		Flickr struct {
			ConsumerKey    string
			ConsumerSecret string
		}
	}

	configPath := flag.String("config", "./config.toml", "")
	flag.Parse()

	if _, err := toml.DecodeFile(*configPath, &conf); err != nil {
		log.Println("ERR decode-config;", err)
		return
	}

	oauthClient := oauth.Client{
		TemporaryCredentialRequestURI: "https://www.flickr.com/services/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://www.flickr.com/services/oauth/authorize",
		TokenRequestURI:               "https://www.flickr.com/services/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  conf.Flickr.ConsumerKey,
			Secret: conf.Flickr.ConsumerSecret,
		},
	}

	var (
		err       error
		tempCred  *oauth.Credentials
		tokenCred *oauth.Credentials
	)

	next := make(chan struct{}, 0)

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenCred, _, err = oauthClient.RequestToken(nil, tempCred, r.FormValue("oauth_verifier"))
			if err != nil {
				log.Fatal(err)
			}

			next <- struct{}{}
		}),
	)
	defer s.Close()

	tempCred, err = oauthClient.RequestTemporaryCredentials(nil, s.URL+"/callback", nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Visit ", oauthClient.AuthorizationURL(tempCred, nil))
	<-next

	log.Printf("%#v", tokenCred)
}
