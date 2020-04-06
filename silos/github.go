package silos

import (
	"context"
	"net/url"
	"regexp"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"hawx.me/code/tally-ho/internal/mfutil"
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

var githubPersonRegexp = regexp.MustCompile(`^https?://github\.com/([^/]+)`)
var githubRepoRegexp = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)`)

func githubParsePersonURL(u string) (username string, ok bool) {
	matches := githubPersonRegexp.FindStringSubmatch(u)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], len(matches[1]) > 0
}

func githubParseRepoURL(u string) (owner, repo string, ok bool) {
	matches := githubRepoRegexp.FindStringSubmatch(u)
	if len(matches) != 3 {
		return "", "", false
	}

	return matches[1], matches[2], true
}

func findGithubRepoURL(vs []interface{}) (u, owner, repo string, ok bool) {
	for _, v := range vs {
		s, ok := v.(string)
		if !ok {
			continue
		}

		owner, repo, ok := githubParseRepoURL(s)
		if !ok {
			continue
		}

		return s, owner, repo, true
	}

	return "", "", "", false
}

func (c *githubClient) Create(data map[string][]interface{}) (location string, err error) {
	switch data["hx-kind"][0].(string) {
	case "like":
		likeOf, owner, repo, ok := findGithubRepoURL(mfutil.GetAll(data, "like-of.properties.url", "like-of"))
		if !ok {
			return "", ErrUnsure{data}
		}

		_, err := c.api.Activity.Star(context.Background(), owner, repo)
		if err != nil {
			return "", err
		}

		return likeOf, nil
	}

	return "", ErrUnsure{data}
}

func (c *githubClient) ResolveCite(u string) (map[string]interface{}, error) {
	owner, repo, ok := githubParseRepoURL(u)
	if !ok {
		return nil, nil
	}

	return map[string]interface{}{
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"url":  {u},
			"name": {owner + "/" + repo},
		},
	}, nil
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
