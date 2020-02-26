package blog

import (
	"database/sql"
	"hawx.me/code/tally-ho/websub"
	"time"
)

type HubStore struct {
	db *sql.DB
}

func NewHubStore(db *sql.DB) (*HubStore, error) {
	s := &HubStore{db}
	return s, s.init()
}

func (h *HubStore) init() error {
	_, err := h.db.Exec(`CREATE TABLE IF NOT EXISTS subscriptions (
    Callback  TEXT,
    Topic     TEXT,
    ExpiresAt DATETIME,
    Secret    TEXT,
    PRIMARY KEY (Callback, Topic)
  );`)

	return err
}

func (h *HubStore) Subscribe(callback, topic string, expiresAt time.Time, secret string) error {
	_, err := h.db.Exec(`
    DELETE FROM subscriptions
      WHERE (Callback = ? AND Topic = ?)
      OR ExpiresAt < ?;

    INSERT INTO subscriptions(Callback, Topic, ExpiresAt, Secret)
      VALUES (?, ?, ?, ?);`,
		callback,
		topic,
		time.Now(),
		callback,
		topic,
		expiresAt,
		secret)

	return err
}

func (h *HubStore) Subscribers(topic string) (websub.SubscribersIter, error) {
	rows, err := h.db.Query(`SELECT Callback, Secret FROM subscriptions WHERE Topic = ? AND ExpiresAt > ?`,
		topic,
		time.Now())
	if err != nil {
		return nil, err
	}

	return &SubscribersIter{rows}, nil
}

func (h *HubStore) Unsubscribe(callback, topic string) error {
	_, err := h.db.Exec(`DELETE FROM subscriptions WHERE Callback = ? AND Topic = ?`,
		callback,
		topic)

	return err
}

type SubscribersIter struct {
	rows *sql.Rows
}

func (s *SubscribersIter) Close() error {
	return s.rows.Close()
}

func (s *SubscribersIter) Data() (callback, secret string, err error) {
	err = s.rows.Scan(&callback, &secret)
	return
}

func (s *SubscribersIter) Err() error {
	return s.rows.Err()
}

func (s *SubscribersIter) Next() bool {
	return s.rows.Next()
}
