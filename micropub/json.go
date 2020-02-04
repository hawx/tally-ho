package micropub

import "strings"

type jsonMicroformat struct {
	Type       []string                 `json:"type,omitempty"`
	Properties map[string][]interface{} `json:"properties"`
	Action     string                   `json:"action,omitempty"`
	URL        string                   `json:"url,omitempty"`
	Add        map[string][]interface{} `json:"add,omitempty"`
	Delete     interface{}              `json:"delete,omitempty"`
	Replace    map[string][]interface{} `json:"replace,omitempty"`
}

func jsonToForm(v jsonMicroformat) map[string][]interface{} {
	if len(v.Type) == 0 {
		v.Type = []string{"h-entry"}
	}

	data := map[string][]interface{}{
		"h": []interface{}{
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

func formToJSON(data map[string][]interface{}) jsonMicroformat {
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
