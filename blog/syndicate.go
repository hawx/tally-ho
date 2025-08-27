package blog

import (
	"log/slog"
)

type Syndicator interface {
	Create(data map[string][]any) (location string, err error)
	UID() string
	Name() string
}

func (b *Blog) syndicate(location string, data map[string][]any) {
	if syndicateTos, ok := data["mp-syndicate-to"]; ok && len(syndicateTos) > 0 {
		for _, syndicateTo := range syndicateTos {
			if syndicator, ok := b.syndicators[syndicateTo.(string)]; ok {
				syndicatedLocation, err := syndicator.Create(data)
				if err != nil {
					slog.Error("create syndication", slog.String("to", syndicator.Name()), slog.Any("uid", data["uid"][0]), slog.Any("err", err))
					continue
				}

				if err := b.Update(location, empty, map[string][]any{
					"syndication": {syndicatedLocation},
				}, empty, []string{}); err != nil {
					slog.Error("confirming syndication", slog.String("to", syndicator.Name()), slog.Any("uid", data["uid"][0]), slog.Any("err", err))
				}
			}
		}
	}
}
