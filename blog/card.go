package blog

import (
	"log/slog"
	"net/http"
	"net/url"

	"willnorris.com/go/microformats"
)

type CardResolver interface {
	ResolveCard(string) (map[string]any, error)
}

func (b *Blog) resolveCard(u string) (map[string]any, error) {
	for _, personer := range b.cardResolvers {
		person, err := personer.ResolveCard(u)
		if err != nil {
			b.logger.Error("resolve card", slog.String("url", u), slog.Any("err", err))
			return nil, nil
		}

		if person == nil {
			continue
		}

		return person, err
	}

	return resolveCard(u)
}

func resolveCard(u string) (card map[string]any, err error) {
	card = map[string]any{
		"type": []any{"h-card"},
		"properties": map[string][]any{
			"url": {u},
		},
	}

	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	uURL, _ := url.Parse(u)
	data := microformats.Parse(resp.Body, uURL)

	for _, item := range data.Items {
		if contains("h-card", item.Type) {
			card = map[string]any{
				"type":       []any{"h-card"},
				"properties": item.Properties,
				"me":         data.Rels["me"],
			}
			return
		}
	}

	return
}
