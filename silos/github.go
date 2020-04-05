package silos

import (
	"context"
	"net/url"
	"regexp"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const GithubUID = "https://github.com/"

type GithubOptions struct {
	BaseURL                string
	ClientID, ClientSecret string
	AccessToken            string
}

func Github(options GithubOptions) (*githubClient, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: options.AccessToken})
	tc := oauth2.NewClient(ctx, ts)
	api := github.NewClient(tc)

	if options.BaseURL != "" {
		baseURL, _ := url.Parse(options.BaseURL)
		api.BaseURL = baseURL
	}

	user, _, err := api.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	return &githubClient{
		api:        api,
		screenName: *user.Login,
	}, nil
}

type githubClient struct {
	api        *github.Client
	screenName string
}

func (c *githubClient) UID() string {
	return GithubUID
}

func (c *githubClient) Name() string {
	return "@" + c.screenName + " on github"
}

var githubPersonRegexp = regexp.MustCompile(`^https?://github\.com/(\w+)`)

func githubParsePersonURL(u string) (username string, ok bool) {
	matches := githubPersonRegexp.FindStringSubmatch(u)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], len(matches[1]) > 0
}

func (c *githubClient) ResolveCard(u string) (map[string]interface{}, error) {
	username, ok := githubParsePersonURL(u)
	if !ok {
		return nil, nil
	}

	return map[string]interface{}{
		"type": []interface{}{"h-card"},
		"properties": map[string][]interface{}{
			"name": {"@" + username},
			"url":  {"https://github.com/" + username},
		},
		"me": []string{"https://github.com/" + username},
	}, nil
}
