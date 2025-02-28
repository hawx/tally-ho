package websub

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
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

type SubscribersIter interface {
	Close() error
	Data() (callback, secret string, err error)
	Err() error
	Next() bool
}

type HubStore interface {
	Subscribe(callback, topic string, expiresAt time.Time, secret string) error
	Subscribers(topic string) (SubscribersIter, error)
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
		baseURL:          baseURL,
		store:            store,
		generator:        challengeGenerator(30),
		noRedirectClient: noRedirectClient,
	}

	return hub
}

type Hub struct {
	baseURL          string
	store            HubStore
	generator        func() ([]byte, error)
	noRedirectClient *http.Client
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "hub.secret must be less than 200 bytes in length", http.StatusBadRequest)
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

	slog.Info("confirm subscription", slog.String("url", callbackURL.String()))

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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	if !bytes.Equal(data, challenge) {
		http.Error(w, "hub.challenge must match", http.StatusBadRequest)
		return
	}

	if err := h.store.Subscribe(callback, topic, time.Now().Add(lease), secret); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	slog.Info("confirmed subscription", slog.String("callback", callback), slog.String("topic", topic), slog.Duration("lease", lease))
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
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	subscribers, err := h.store.Subscribers(topic)
	if err != nil {
		return err
	}
	defer subscribers.Close()

	client := h.noRedirectClient
	link := `<` + h.baseURL + `>; rel="hub", <` + topic + `>; rel="self"`

	for subscribers.Next() {
		callback, secret, err := subscribers.Data()
		if err != nil {
			continue
		}

		req, err := http.NewRequest("POST", callback, bytes.NewReader(body))
		if err != nil {
			continue
		}

		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Link", link)

		if secret != "" {
			mac := hmac.New(sha512.New, []byte(secret))
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
			h.store.Unsubscribe(callback, topic)
		}
	}

	return subscribers.Err()
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
