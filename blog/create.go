package blog

import (
	"log"
	"net/url"
)

func (b *Blog) Create(data map[string][]interface{}) (location string, err error) {
	if len(data["name"]) == 0 {
		name, err := getName(data)
		if err != nil {
			log.Printf("WARN get-name; %v\n", err)
		} else if name != "" {
			data["name"] = []interface{}{name}
			log.Printf("WARN get-name; setting to '%s'\n", name)
		}
	}

	relativeLocation, err := b.DB.Create(data)
	if err != nil {
		return
	}

	relativeURL, _ := url.Parse(relativeLocation)
	location = b.Config.BaseURL.ResolveReference(relativeURL).String()

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
				}, empty); err != nil {
					log.Printf("ERR confirming-syndication to=%s uid=%s; %v\n", syndicator.Name(), data["uid"][0], err)
				}
			}
		}
	}

	return
}