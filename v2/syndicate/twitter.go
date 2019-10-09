package syndicate

import (
	"errors"
	"net/url"

	"github.com/ChimeraCoder/anaconda"
)

var ErrUnsure = errors.New("unsure what to create")

type TwitterOptions struct {
	BaseURL                        string
	ConsumerKey, ConsumerSecret    string
	AccessToken, AccessTokenSecret string
}

func Twitter(options TwitterOptions) *twitterSyndicator {
	api := anaconda.NewTwitterApiWithCredentials(
		options.AccessToken,
		options.AccessTokenSecret,
		options.ConsumerKey,
		options.ConsumerSecret,
	)

	if options.BaseURL != "" {
		api.SetBaseUrl(options.BaseURL)
	}

	return &twitterSyndicator{
		api: api,
	}
}

type twitterSyndicator struct {
	api *anaconda.TwitterApi
}

func (t *twitterSyndicator) Create(data map[string][]interface{}) (location string, err error) {
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
