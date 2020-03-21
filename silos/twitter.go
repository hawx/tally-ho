package silos

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/ChimeraCoder/anaconda"
	"hawx.me/code/tally-ho/internal/mfutil"
	"mvdan.cc/xurls/v2"
)

// TwitterUID is the unique identifier for twitter.
const TwitterUID = "https://twitter.com/"

// TwitterOptions is the configuration required to connect to the Twitter API.
type TwitterOptions struct {
	BaseURL                        string
	ConsumerKey, ConsumerSecret    string
	AccessToken, AccessTokenSecret string
}

// Twitter creates a client for Twitter. On creation it makes a call to the
// API to verify the credentials are correct and the screen name of the
// authenticated user.
func Twitter(options TwitterOptions) (*twitterClient, error) {
	api := anaconda.NewTwitterApiWithCredentials(
		options.AccessToken,
		options.AccessTokenSecret,
		options.ConsumerKey,
		options.ConsumerSecret,
	)

	if options.BaseURL != "" {
		api.SetBaseUrl(options.BaseURL)
	}

	user, err := api.GetSelf(url.Values{})
	if err != nil {
		return nil, err
	}

	return &twitterClient{
		api:        api,
		screenName: user.ScreenName,
	}, nil
}

type twitterClient struct {
	api        *anaconda.TwitterApi
	screenName string
}

func (t *twitterClient) UID() string {
	return TwitterUID
}

func (t *twitterClient) Name() string {
	return "@" + t.screenName + " on twitter"
}

var twitterStatusRegexp = regexp.MustCompile(`^https?://twitter\.com/(?:\#!/)?(\w+)/status(es)?/(\d+)`)
var twitterPersonRegexp = regexp.MustCompile(`^https?://twitter\.com/(?:\#!/)?(\w+)`)

func twitterParseStatusURL(u string) (tweetID int64, username string, ok bool) {
	matches := twitterStatusRegexp.FindStringSubmatch(u)
	if len(matches) != 4 {
		return 0, "", false
	}

	tweetID, err := strconv.ParseInt(matches[3], 10, 0)

	return tweetID, matches[1], err == nil
}

func findTwitterStatusURL(vs []interface{}) (url string, tweetID int64, username string, ok bool) {
	for _, v := range vs {
		s, ok := v.(string)
		if !ok {
			continue
		}

		tweetID, username, ok := twitterParseStatusURL(s)
		if !ok {
			continue
		}

		return s, tweetID, username, ok
	}

	return "", 0, "", false
}

func twitterParsePersonURL(u string) (username string, ok bool) {
	matches := twitterPersonRegexp.FindStringSubmatch(u)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], len(matches[1]) > 0
}

func (t *twitterClient) Create(data map[string][]interface{}) (location string, err error) {
	switch data["hx-kind"][0].(string) {
	case "like":
		likeOf, tweetID, _, ok := findTwitterStatusURL(
			mfutil.GetAll(data, "like-of.properties.url", "like-of"),
		)
		if !ok {
			return "", ErrUnsure{data}
		}

		_, err := t.api.Favorite(tweetID)
		if err != nil {
			return "", err
		}

		return likeOf, nil

	case "reply":
		_, tweetID, username, ok := findTwitterStatusURL(
			mfutil.GetAll(data, "in-reply-to.properties.url", "in-reply-to"),
		)
		if !ok {
			return "", ErrUnsure{data}
		}

		content, ok := autoLinkContent(data)
		if !ok {
			return "", ErrUnsure{data}
		}

		tweet, err := t.api.PostTweet("@"+username+" "+content, url.Values{
			"in_reply_to_status_id": {strconv.FormatInt(tweetID, 10)},
		})
		if err != nil {
			return "", err
		}

		return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil

	case "photo":
		photos, ok := data["photo"]
		if !ok {
			return "", ErrUnsure{data}
		}

		var mediaIDs []string
		for _, photo := range photos {
			var photoURL string
			if u, ok := photo.(string); ok {
				photoURL = u
			} else if m, ok := photo.(map[string]interface{}); ok {
				if u, ok := m["value"].(string); ok {
					photoURL = u
				} else {
					continue
				}
			} else {
				continue
			}

			resp, err := http.Get(photoURL)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			var buf bytes.Buffer
			enc := base64.NewEncoder(base64.StdEncoding, &buf)
			if _, err := io.Copy(enc, resp.Body); err != nil {
				return "", err
			}
			if err := enc.Close(); err != nil {
				return "", err
			}

			media, err := t.api.UploadMedia(buf.String())
			if err != nil {
				return "", err
			}

			mediaIDs = append(mediaIDs, media.MediaIDString)
		}

		content, ok := autoLinkContent(data)
		if !ok {
			content = ""
		}

		tweet, err := t.api.PostTweet(content, url.Values{
			"media_ids": {strings.Join(mediaIDs, ",")},
		})
		if err != nil {
			return "", err
		}

		return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil

	case "note":
		content, ok := autoLinkContent(data)
		if !ok {
			return "", ErrUnsure{data}
		}

		tweet, err := t.api.PostTweet(content, url.Values{})
		if err != nil {
			return "", err
		}

		return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil

	case "article":
		name, ok := mfutil.Get(data, "name").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		u, ok := mfutil.Get(data, "url").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		tweet, err := t.api.PostTweet(name+" † "+u, url.Values{})
		if err != nil {
			return "", err
		}

		return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil
	}

	return "", ErrUnsure{data}
}

func autoLinkContent(data map[string][]interface{}) (string, bool) {
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

func (t *twitterClient) ResolveCite(u string) (map[string]interface{}, error) {
	tweetID, _, ok := twitterParseStatusURL(u)
	if !ok {
		return nil, nil
	}

	tweet, err := t.api.GetTweet(tweetID, url.Values{})

	return map[string]interface{}{
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"name":    {"@" + tweet.User.ScreenName + "'s tweet"},
			"content": {tweet.FullText},
			"url":     {u},
			"author": {
				map[string]interface{}{
					"type": []interface{}{"h-card"},
					"properties": map[string][]interface{}{
						"name":     {tweet.User.Name},
						"url":      {"https://twitter.com/" + tweet.User.ScreenName},
						"nickname": {"@" + tweet.User.ScreenName},
					},
				},
			},
		},
	}, err
}

// ResolveCard attempts to resolve a given URL to a Twitter profile, it does
// this by checking if the URL matches a regexp. It will return a h-card with
// name='@screename', because that is how people are referred to in tweets even
// though it would be "more correct" to return this as nickname.
func (t *twitterClient) ResolveCard(u string) (map[string]interface{}, error) {
	username, ok := twitterParsePersonURL(u)
	if !ok {
		return nil, nil
	}

	return map[string]interface{}{
		"type": []interface{}{"h-card"},
		"properties": map[string][]interface{}{
			"name": {"@" + username},
			"url":  {"https://twitter.com/" + username},
		},
		"me": []string{"https://twitter.com/" + username},
	}, nil
}
