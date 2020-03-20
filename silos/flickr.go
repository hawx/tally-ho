package silos

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gomodule/oauth1/oauth"
	"hawx.me/code/tally-ho/internal/mfutil"
)

const flickrBaseURL = "https://www.flickr.com/services/rest"

const FlickrUID = "https://flickr.com/"

type FlickrOptions struct {
	BaseURL                        string
	ConsumerKey, ConsumerSecret    string
	AccessToken, AccessTokenSecret string
}

func Flickr(options FlickrOptions) (*flickrClient, error) {
	client := &flickrClient{
		baseURL: flickrBaseURL,
		client:  http.DefaultClient,
		credentials: &oauth.Credentials{
			Token:  options.AccessToken,
			Secret: options.AccessTokenSecret,
		},
	}

	oauthClient := &oauth.Client{
		TemporaryCredentialRequestURI: "https://www.flickr.com/services/oauth/request_token",
		ResourceOwnerAuthorizationURI: "https://www.flickr.com/services/oauth/authorize",
		TokenRequestURI:               "https://www.flickr.com/services/oauth/access_token",
		Credentials: oauth.Credentials{
			Token:  options.ConsumerKey,
			Secret: options.ConsumerSecret,
		},
	}

	if options.BaseURL != "" {
		client.baseURL = options.BaseURL
	}

	resp, err := oauthClient.Get(client.client, client.credentials, client.baseURL, url.Values{
		"format":         {"json"},
		"nojsoncallback": {"1"},
		"method":         {"flickr.test.login"},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var v struct {
		User struct {
			Username struct {
				Content string `json:"_content"`
			} `json:"username"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	client.oauthClient = oauthClient
	client.screenName = v.User.Username.Content

	return client, nil
}

type flickrClient struct {
	baseURL     string
	client      *http.Client
	oauthClient *oauth.Client
	credentials *oauth.Credentials
	screenName  string
}

func (c *flickrClient) get(method string, v interface{}) error {
	qs := url.Values{
		"nojsoncallback": {"1"},
		"format":         {"json"},
		"method":         {method},
	}

	resp, err := c.client.Get(c.baseURL + qs.Encode())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&v)
}

func (*flickrClient) UID() string {
	return FlickrUID
}

func (c *flickrClient) Name() string {
	return c.screenName + " on flickr"
}

var flickrPhotoRegexp = regexp.MustCompile(`^https?://www\.flickr\.com/photos/[a-zA-Z\-]+/(\d+)/`)

func flickrParseURL(u string) (photoID string, ok bool) {
	matches := flickrPhotoRegexp.FindStringSubmatch(u)
	if len(matches) != 2 {
		return "", false
	}

	return matches[1], true
}

func (c *flickrClient) Create(data map[string][]interface{}) (location string, err error) {
	switch data["hx-kind"][0].(string) {
	case "like":
		likeOf, ok := mfutil.Get(data, "like-of.properties.url", "like-of").(string)
		if !ok {
			return "", ErrUnsure{data}
		}

		if photoID, ok := flickrParseURL(likeOf); ok {
			resp, err := c.oauthClient.Post(c.client, c.credentials, c.baseURL, url.Values{
				"method":   {"flickr.favorites.add"},
				"photo_id": {photoID},
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return "", errors.New("flickr add new comment got: " + resp.Status)
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

		if photoID, ok := flickrParseURL(replyTo); ok {
			resp, err := c.oauthClient.Post(c.client, c.credentials, c.baseURL, url.Values{
				"format":         {"json"},
				"nojsoncallback": {"1"},
				"method":         {"flickr.photos.comments.addComment"},
				"photo_id":       {photoID},
				"comment_text":   {content},
			})
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				return "", errors.New("flickr add new comment got: " + resp.Status)
			}

			var v struct {
				Comment struct {
					Permalink string `json:"permalink"`
				} `json:"comment"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
				return "", err
			}

			return v.Comment.Permalink, nil
		}
	}

	return "", ErrUnsure{data}
}

func (c *flickrClient) ResolveCite(u string) (map[string]interface{}, error) {
	photoID, ok := flickrParseURL(u)
	if !ok {
		return nil, nil
	}

	resp, err := c.oauthClient.Get(c.client, c.credentials, c.baseURL, url.Values{
		"format":         {"json"},
		"nojsoncallback": {"1"},
		"method":         {"flickr.photos.getInfo"},
		"photo_id":       {photoID},
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var v struct {
		Photo struct {
			ID    string `json:"id"`
			Owner struct {
				NSID     string `json:"nsid"`
				Username string `json:"username"`
				Name     string `json:"realname"`
			} `json:"owner"`
			Title struct {
				Content string `json:"_content"`
			} `json:"title"`
			Description struct {
				Content string `json:"_content"`
			} `json:"description"`
			Visibility struct {
				IsPublic int `json:"ispublic"`
			} `json:"visibility"`
		} `json:"photo"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	if v.Photo.Visibility.IsPublic != 1 {
		return nil, nil
	}

	resp2, err := c.oauthClient.Get(c.client, c.credentials, c.baseURL, url.Values{
		"format":         {"json"},
		"nojsoncallback": {"1"},
		"method":         {"flickr.photos.getSizes"},
		"photo_id":       {photoID},
	})
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var vv struct {
		Sizes struct {
			Size []struct {
				Label  string `json:"label"`
				Width  int
				Source string `json:"source"`
			} `json:"size"`
		} `json:"sizes"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&vv); err != nil {
		return nil, err
	}

	authorProperties := map[string][]interface{}{
		"url": {"https://www.flickr.com/" + v.Photo.Owner.NSID},
	}

	if v.Photo.Owner.Name != "" {
		authorProperties["name"] = []interface{}{v.Photo.Owner.Name}
	}

	if v.Photo.Owner.Username != "" {
		authorProperties["nickname"] = []interface{}{v.Photo.Owner.Username}
	}

	properties := map[string][]interface{}{
		"url": {u},
		"author": {
			map[string]interface{}{
				"type":       []interface{}{"h-card"},
				"properties": authorProperties,
			},
		},
	}

	properties["name"] = []interface{}{v.Photo.Owner.Username + "'s photo"}
	if v.Photo.Title.Content != "" {
		properties["name"][0] = v.Photo.Title.Content
	}

	if v.Photo.Description.Content != "" {
		properties["content"] = []interface{}{v.Photo.Description.Content}
	}

	for _, size := range vv.Sizes.Size {
		properties["photo"] = []interface{}{size.Source}
		if size.Width > 768 {
			break
		}
	}

	return map[string]interface{}{
		"type":       []interface{}{"h-cite"},
		"properties": properties,
	}, nil
}
