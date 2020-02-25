package silos

import (
	"net/url"
	"regexp"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
	"hawx.me/code/tally-ho/internal/mfutil"
)

// TwitterUID is the unique identifier for this syndicator.
const TwitterUID = "https://twitter.com/"

// TwitterOptions is the configuration required to connect to the Twitter API.
type TwitterOptions struct {
	BaseURL                        string
	ConsumerKey, ConsumerSecret    string
	AccessToken, AccessTokenSecret string
}

// Twitter creates a syndicator for Twitter. On creation it makes a call to the
// API to verify the credentials are correct and the screen name of the
// authenticated user.
func Twitter(options TwitterOptions) (*twitterSyndicator, error) {
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

	return &twitterSyndicator{
		api:        api,
		screenName: user.ScreenName,
	}, nil
}

type twitterSyndicator struct {
	api        *anaconda.TwitterApi
	screenName string
}

func (t *twitterSyndicator) UID() string {
	return TwitterUID
}

func (t *twitterSyndicator) Name() string {
	return "@" + t.screenName + " on twitter"
}

var twitterStatusRegexp = regexp.MustCompile(`^https?://twitter\.com/(?:\#!/)?(\w+)/status(es)?/(\d+)`)

func twitterParseURL(u string) (tweetID int64, username string, ok bool) {
	matches := twitterStatusRegexp.FindStringSubmatch(u)
	if len(matches) != 4 {
		return 0, "", false
	}

	tweetID, err := strconv.ParseInt(matches[3], 10, 0)

	return tweetID, matches[1], err == nil
}

func (t *twitterSyndicator) Create(data map[string][]interface{}) (location string, err error) {
	switch data["hx-kind"][0].(string) {
	case "like":
		likeOf, ok := mfutil.Get(data, "like-of.properties.url", "like-of").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		if tweetID, _, ok := twitterParseURL(likeOf); ok {
			_, err := t.api.Favorite(tweetID)
			if err != nil {
				return "", err
			}

			return likeOf, nil
		}

	case "reply":
		replyTo, ok := mfutil.Get(data, "in-reply-to.properties.url", "in-reply-to").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		content, ok := mfutil.Get(data, "content.text", "content").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		if tweetID, username, ok := twitterParseURL(replyTo); ok {
			tweet, err := t.api.PostTweet("@"+username+" "+content, url.Values{
				"in_reply_to_status_id": {strconv.FormatInt(tweetID, 10)},
			})
			if err != nil {
				return "", err
			}

			return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil
		}

	case "note":
		content, ok := mfutil.Get(data, "content.text", "content").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		tweet, err := t.api.PostTweet(content, url.Values{})
		if err != nil {
			return "", err
		}

		return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil
	}

	return "", ErrUnsure{data}
}

func (t *twitterSyndicator) Cite(u string) (map[string]interface{}, error) {
	tweetID, _, ok := twitterParseURL(u)
	if !ok {
		return nil, nil
	}

	tweet, err := t.api.GetTweet(tweetID, url.Values{})

	return map[string]interface{}{
		"type": []interface{}{"h-cite"},
		"properties": map[string][]interface{}{
			"name":    {"@" + tweet.User.ScreenName + "'s tweet"},
			"content": {tweet.FullText},
			"author": {
				map[string]interface{}{
					"type": []interface{}{"h-card"},
					"properties": map[string][]interface{}{
						"name":     {tweet.User.Name},
						"url":      {tweet.User.URL},
						"nickname": {"@" + tweet.User.ScreenName},
					},
				},
			},
		},
	}, err
}
