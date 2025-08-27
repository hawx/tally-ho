package blog

import (
	"errors"
	"log/slog"
	"slices"

	"hawx.me/code/numbersix"
	"hawx.me/code/tally-ho/internal/mfutil"
)

func (b *Blog) Create(data map[string][]any) (string, error) {
	b.massage(data)

	uid := mfutil.Get(data, "uid").(string)
	location := mfutil.Get(data, "url").(string)

	triples, err := b.entries.List(numbersix.Where("uid", uid))
	if err != nil {
		return location, err
	}

	if len(data["hx-url"]) == 0 {
		if len(triples) > 0 {
			return location, errors.New("post with uid already exists")
		}
	} else {
		if err := b.entries.DeleteSubject(uid); err != nil {
			return location, err
		}
	}

	if err := b.entries.SetProperties(uid, data); err != nil {
		return location, err
	}

	slog.Info("set entry properties", slog.String("uid", uid), slog.String("url", location))

	// don't send updates for published pages (although maybe I'll change this in the future)
	if len(data["hx-url"]) == 0 {
		go b.syndicate(location, data)
		go b.sendWebmentions(location, data)
		go b.hubPublish()
	}

	return location, nil
}

func contains(needle string, list []string) bool {
	return slices.Contains(list, needle)
}
