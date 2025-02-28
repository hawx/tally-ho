package blog

import (
	"log/slog"
	"net/url"
	"time"
)

type HubPublisher interface {
	Publish(topic string) error
}

func (b *Blog) hubPublish() {
	// for now just publish the main things that will always change
	changed := []string{
		"/",
		"/feed/atom",
		"/feed/jsonfeed",
		"/feed/rss",
	}

	// ensure that the entry exists
	time.Sleep(time.Second)

	for _, c := range changed {
		u, err := url.Parse(c)
		if err != nil {
			b.logger.Warn("hub publish parse url", slog.String("url", c), slog.Any("err", err))
			continue
		}

		err = b.hubPublisher.Publish(b.config.BaseURL.ResolveReference(u).String())
		if err != nil {
			b.logger.Warn("hub publish", slog.String("url", c), slog.Any("err", err))
		}
	}
}
