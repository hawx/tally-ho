package websub

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultLease = 28 * 24 * time.Hour

type HubStore interface {
	Subscribe(callback, topic string, expiresAt time.Time) error
	Subscribers(topic string) ([]string, error)
	Unsubscribe(callback, topic string) error
}

func New(store HubStore) *Hub {
	hub := &Hub{Store: store, generator: challengeGenerator(30)}

	return hub
}

type Hub struct {
	Store     HubStore
	generator func() ([]byte, error)
}

func (h *Hub) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST is allowed", http.StatusMethodNotAllowed)
			return
		}

		var (
			callback = r.FormValue("hub.callback")
			mode     = r.FormValue("hub.mode")
			topic    = r.FormValue("hub.topic")
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

		challenge, err := h.generator()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		query := callbackURL.Query()
		query.Add("hub.mode", mode)
		query.Add("hub.topic", topic)
		query.Add("hub.challenge", string(challenge))
		query.Add("hub.lease_seconds", strconv.Itoa(int(defaultLease.Seconds())))
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

		h.Store.Subscribe(callback, topic, time.Now().Add(defaultLease))
		w.WriteHeader(http.StatusAccepted)
	})

	return mux
}

func (h *Hub) Publish(topic string) error {
	resp, err := http.Get(topic)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// todo
	}

	contentType := resp.Header.Get("Content-Type")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	subscribers, err := h.Store.Subscribers(topic)
	if err != nil {
		// todo
	}

	// todo: no redirect client
	client := http.DefaultClient
	link := `<the-hub-url>; rel="hub", <` + topic + `>; rel="self"`

	for _, subscriber := range subscribers {
		req, err := http.NewRequest("POST", subscriber, bytes.NewReader(body))
		if err != nil {
			continue
		}

		req.Header.Add("Content-Type", contentType)
		req.Header.Add("Link", link)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusGone {
			h.Store.Unsubscribe(subscriber, topic)
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
