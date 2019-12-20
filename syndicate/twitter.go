package syndicate

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
)

var ErrUnsure = errors.New("unsure what to create")

type TwitterOptions struct {
	BaseURL                        string
	ConsumerKey, ConsumerSecret    string
	AccessToken, AccessTokenSecret string
}

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

func (t *twitterSyndicator) Config() Config {
	return Config{
		UID:  "https://twitter.com/",
		Name: "@" + t.screenName + " on twitter",
	}
}

var twitterStatusRegexp = regexp.MustCompile(`^https?://twitter\.com/(?:\#!/)?\w+/status(es)?/(\d+)$`)

func (t *twitterSyndicator) Create(data map[string][]interface{}) (location string, err error) {
	if len(data["like-of"]) > 0 {
		likeOf, ok := data["like-of"][0].(string)
		if !ok {
			return "", ErrUnsure
		}

		matches := twitterStatusRegexp.FindStringSubmatch(likeOf)
		if len(matches) == 3 {
			tweetID, err := strconv.ParseInt(matches[2], 10, 0)
			if err == nil {
				_, err := t.api.Favorite(tweetID)
				if err != nil {
					return "", err
				}

				return likeOf, nil
			}
		}
	}

	contents := data["content"]
	if len(contents) < 1 {
		return "", ErrUnsure
	}

	content, ok := contents[0].(string)
	if !ok {
		return "", ErrUnsure
	}

	tweet, err := t.api.PostTweet(content, url.Values{})
	if err != nil {
		return "", err
	}

	return "https://twitter.com/" + tweet.User.ScreenName + "/status/" + tweet.IdStr, nil
}
