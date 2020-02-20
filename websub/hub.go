package websub

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultLease = 10 * 24 * time.Hour
	maxLease     = 28 * 24 * time.Hour
)

type Subscriber struct {
	Callback string
	Secret   string
}

type HubStore interface {
	Subscribe(callback, topic string, expiresAt time.Time, secret string) error
	Subscribers(topic string) ([]Subscriber, error)
	Unsubscribe(callback, topic string) error
}

func New(baseURL string, store HubStore) *Hub {
	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 10 * time.Second,
	}

	hub := &Hub{
		BaseURL:          baseURL,
		Store:            store,
		generator:        challengeGenerator(30),
		noRedirectClient: noRedirectClient,
	}

	return hub
}

type Hub struct {
	BaseURL          string
	Store            HubStore
	generator        func() ([]byte, error)
	noRedirectClient *http.Client
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	var (
		callback     = r.FormValue("hub.callback")
		mode         = r.FormValue("hub.mode")
		topic        = r.FormValue("hub.topic")
		leaseSeconds = r.FormValue("hub.lease_seconds")
		secret       = r.FormValue("hub.secret")
	)

	callbackURL, err := url.Parse(callback)
	if err != nil || !callbackURL.IsAbs() {
		http.Error(w, "hub.callback must be a url", http.StatusBadRequest)
		return
	}

	if mode != "subscribe" && mode != "unsubscribe" {
		http.Error(w, "hub.mode must be subscribe or unsubscribe", http.StatusBadRequest)
		return
	}

	if len(secret) > 200 {
		http.Error(w, "hub.secret must be less than 200i bytes in length", http.StatusBadRequest)
		return
	}

	challenge, err := h.generator()
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	lease := defaultLease
	if leaseSeconds != "" {
		if l, err := strconv.Atoi(leaseSeconds); err == nil {
			lease = time.Duration(l) * time.Second
		}

		if lease > maxLease {
			lease = maxLease
		}
	}

	query := callbackURL.Query()
	query.Add("hub.mode", mode)
	query.Add("hub.topic", topic)
	query.Add("hub.challenge", string(challenge))
	query.Add("hub.lease_seconds", strconv.Itoa(int(lease.Seconds())))
	callbackURL.RawQuery = query.Encode()

	resp, err := http.Get(callbackURL.String())
	if err != nil {
		http.Error(w, "problem requesting hub.callback", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		http.Error(w, "hub.callback returned a non-200 response", http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	if !bytes.Equal(data, challenge) {
		http.Error(w, "hub.challenge must match", http.StatusBadRequest)
		return
	}

	h.Store.Subscribe(callback, topic, time.Now().Add(lease), secret)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Hub) Publish(topic string) error {
	resp, err := http.Get(topic)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("could not retrieve topic: " + topic)
	}

	contentType := resp.Header.Get("Content-Type")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	subscribers, err := h.Store.Subscribers(topic)
	if err != nil {
		return err
	}

	client := h.noRedirectClient
	link := `<` + h.BaseURL + `>; rel="hub", <` + topic + `>; rel="self"`

	for _, subscriber := range subscribers {
		req, err := http.NewRequest("POST", subscriber.Callback, bytes.NewReader(body))
		if err != nil {
			continue
		}

		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Link", link)

		if subscriber.Secret != "" {
			mac := hmac.New(sha512.New, []byte(subscriber.Secret))
			if _, err := mac.Write(body); err != nil {
				continue
			}
			req.Header.Add("X-Hub-Signature", "sha512="+hex.EncodeToString(mac.Sum(nil)))
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusGone {
			h.Store.Unsubscribe(subscriber.Callback, topic)
		}
	}

	return nil
}

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

func challengeGenerator(n int) func() ([]byte, error) {
	return func() ([]byte, error) {
		bytes := make([]byte, n)
		_, err := rand.Read(bytes)
		if err != nil {
			return []byte{}, err
		}
		for i, b := range bytes {
			bytes[i] = letters[b%byte(len(letters))]
		}
		return bytes, nil
	}
}
