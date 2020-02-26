package blog

import (
	"database/sql"
	"testing"
	"time"

	"hawx.me/code/assert"
)

func TestHubStoreSubscribe(t *testing.T) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(err)

	store, err := NewHubStore(db)
	assert.Nil(err)

	err = store.Subscribe("callback", "topic", time.Now().Add(5*time.Second), "secret")
	assert.Nil(err)

	subscribers, err := store.Subscribers("topic")
	assert.Nil(err)
	assert.True(subscribers.Next())

	callback, secret, err := subscribers.Data()
	assert.Equal("callback", callback)
	assert.Equal("secret", secret)
	assert.Nil(err)

	assert.False(subscribers.Next())
	assert.Nil(subscribers.Err())
	assert.Nil(subscribers.Close())
}

func TestHubStoreSubscribeWhenExpired(t *testing.T) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(err)

	store, err := NewHubStore(db)
	assert.Nil(err)

	err = store.Subscribe("callback", "topic", time.Now().Add(-5*time.Second), "secret")
	assert.Nil(err)

	subscribers, err := store.Subscribers("topic")
	assert.Nil(err)

	assert.False(subscribers.Next())
	assert.Nil(subscribers.Err())
	assert.Nil(subscribers.Close())
}

func TestHubStoreSubscribeTwice(t *testing.T) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(err)

	store, err := NewHubStore(db)
	assert.Nil(err)

	err = store.Subscribe("callback", "topic", time.Now().Add(5*time.Second), "secret")
	assert.Nil(err)
	err = store.Subscribe("callback", "topic", time.Now().Add(10*time.Second), "newsecret")
	assert.Nil(err)

	subscribers, err := store.Subscribers("topic")
	assert.Nil(err)
	assert.True(subscribers.Next())

	callback, secret, err := subscribers.Data()
	assert.Equal("callback", callback)
	assert.Equal("newsecret", secret)
	assert.Nil(err)

	assert.False(subscribers.Next())
	assert.Nil(subscribers.Err())
	assert.Nil(subscribers.Close())
}

func TestHubStoreUnsubscribe(t *testing.T) {
	assert := assert.New(t)

	db, err := sql.Open("sqlite3", ":memory:")
	assert.Nil(err)

	store, err := NewHubStore(db)
	assert.Nil(err)

	err = store.Subscribe("callback", "topic", time.Now().Add(5*time.Second), "secret")
	assert.Nil(err)
	err = store.Unsubscribe("callback", "topic")
	assert.Nil(err)

	subscribers, err := store.Subscribers("topic")
	assert.Nil(err)

	assert.False(subscribers.Next())
	assert.Nil(subscribers.Err())
	assert.Nil(subscribers.Close())
}
