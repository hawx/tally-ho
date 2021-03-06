package silos

import (
	"context"
	"net/url"
	"regexp"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"hawx.me/code/tally-ho/internal/mfutil"
	"mvdan.cc/xurls/v2"
)

const GithubUID = "https://github.com/"

type GithubOptions struct {
	BaseURL     string
	AccessToken string
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
var githubIssuesRegexp = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/issues/(\d+)`)

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

func githubParseIssuesURL(u string) (owner, repo string, issue int, ok bool) {
	matches := githubIssuesRegexp.FindStringSubmatch(u)
	if len(matches) != 4 {
		return "", "", 0, false
	}

	issue, err := strconv.Atoi(matches[3])
	if err != nil {
		return "", "", 0, false
	}

	return matches[1], matches[2], issue, true
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

func findGithubIssuesURL(vs []interface{}) (u, owner, repo string, issue int, ok bool) {
	for _, v := range vs {
		s, ok := v.(string)
		if !ok {
			continue
		}

		owner, repo, issue, ok := githubParseIssuesURL(s)
		if !ok {
			continue
		}

		return s, owner, repo, issue, true
	}

	return "", "", "", 0, false
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

	case "reply":
		if _, owner, repo, issue, ok := findGithubIssuesURL(mfutil.GetAll(data, "in-reply-to.properties.url", "in-reply-to")); ok {
			req := &github.IssueComment{}
			if content, ok := githubAutoLinkContent(data); ok {
				req.Body = &content
			}

			issue, _, err := c.api.Issues.CreateComment(context.Background(), owner, repo, issue, req)
			if err != nil {
				return "", err
			}

			return *issue.HTMLURL, nil
		}

		if _, owner, repo, ok := findGithubRepoURL(mfutil.GetAll(data, "in-reply-to.properties.url", "in-reply-to")); ok {
			req := &github.IssueRequest{}
			if name, ok := mfutil.Get(data, "name").(string); ok {
				req.Title = &name
			}
			if content, ok := githubAutoLinkContent(data); ok {
				req.Body = &content
			}

			issue, _, err := c.api.Issues.Create(context.Background(), owner, repo, req)
			if err != nil {
				return "", err
			}

			return *issue.HTMLURL, nil
		}
	}

	return "", ErrUnsure{data}
}

func githubAutoLinkContent(data map[string][]interface{}) (string, bool) {
	content, ok := mfutil.Get(data, "content.text", "content").(string)
	if !ok {
		return "", false
	}

	people, ok := mfutil.Get(data, "hx-people").(map[string][]string)
	if !ok {
		people = map[string][]string{}
	}
	reg := xurls.Strict()

	content = regexp.
		MustCompile("@"+reg.String()).
		ReplaceAllStringFunc(content, func(u string) string {
			if found, ok := people[u[1:]]; ok {
				for _, u := range found {
					if username, ok := twitterParsePersonURL(u); ok {
						return "@" + username
					}
				}
			}

			return u
		})

	return content, true
}

func (c *githubClient) ResolveCite(u string) (map[string]interface{}, error) {
	if owner, repo, issue, ok := githubParseIssuesURL(u); ok {
		return map[string]interface{}{
			"type": []interface{}{"h-cite"},
			"properties": map[string][]interface{}{
				"url":  {u},
				"name": {owner + "/" + repo + "#" + strconv.Itoa(issue)},
			},
		}, nil
	}

	if owner, repo, ok := githubParseRepoURL(u); ok {
		return map[string]interface{}{
			"type": []interface{}{"h-cite"},
			"properties": map[string][]interface{}{
				"url":  {u},
				"name": {owner + "/" + repo},
			},
		}, nil
	}

	return nil, nil
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
