package blog

import (
	"net/http"
	"net/url"

	"willnorris.com/go/microformats"
)

type Personer interface {
	Person(string) (map[string]interface{}, error)
}

func (b *Blog) getPerson(u string) (map[string]interface{}, error) {
	// for _, personer := range b.personers {
	// 	person, err := personer.Person(u)
	// 	if err != nil {
	// 		log.Printf("ERR get-person url=%s; %v\n", u, err)
	// 		return nil, nil
	// 	}

	// 	if person == nil {
	// 		continue
	// 	}

	// 	return person, err
	// }

	return getPerson(u)
}

func getPerson(u string) (card map[string]interface{}, err error) {
	card = map[string]interface{}{
		"type": []interface{}{"h-card"},
		"properties": map[string][]interface{}{
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
			card = map[string]interface{}{
				"type":       []interface{}{"h-card"},
				"properties": item.Properties,
				"me":         data.Rels["me"],
			}
			return
		}
	}

	return
}
