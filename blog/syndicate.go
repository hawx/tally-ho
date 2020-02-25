package blog

import "log"

type Syndicator interface {
	Create(data map[string][]interface{}) (location string, err error)
	UID() string
	Name() string
}

func (b *Blog) syndicate(location string, data map[string][]interface{}) {
	if syndicateTos, ok := data["mp-syndicate-to"]; ok && len(syndicateTos) > 0 {
		for _, syndicateTo := range syndicateTos {
			if syndicator, ok := b.Syndicators[syndicateTo.(string)]; ok {
				syndicatedLocation, err := syndicator.Create(data)
				if err != nil {
					log.Printf("ERR syndication to=%s uid=%s; %v\n", syndicator.Name(), data["uid"][0], err)
					continue
				}

				if err := b.Update(location, empty, map[string][]interface{}{
					"syndication": {syndicatedLocation},
				}, empty, []string{}); err != nil {
					log.Printf("ERR confirming-syndication to=%s uid=%s; %v\n", syndicator.Name(), data["uid"][0], err)
				}
			}
		}
	}
}
