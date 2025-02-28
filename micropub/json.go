package micropub

import "strings"

type jsonMicroformat struct {
	Type       []string         `json:"type,omitempty"`
	Properties map[string][]any `json:"properties"`
	Action     string           `json:"action,omitempty"`
	URL        string           `json:"url,omitempty"`
	Add        map[string][]any `json:"add,omitempty"`
	Delete     any              `json:"delete,omitempty"`
	Replace    map[string][]any `json:"replace,omitempty"`
}

func jsonToForm(v jsonMicroformat) map[string][]any {
	if len(v.Type) == 0 {
		v.Type = []string{"h-entry"}
	}

	data := map[string][]any{
		"h": {
			strings.TrimPrefix(v.Type[0], "h-"),
		},
	}

	for key, value := range v.Properties {
		if reservedKey(key) || len(value) == 0 {
			continue
		}

		data[key] = value
	}

	return data
}

func formToJSON(data map[string][]any) jsonMicroformat {
	var htype []string
	if len(data["h"]) == 1 {
		htype = []string{"h-" + data["h"][0].(string)}
		delete(data, "h")
	}

	return jsonMicroformat{
		Type:       htype,
		Properties: data,
	}
}
