package blog

import (
	"html/template"
	"path/filepath"
	"strings"
)

func ParseTemplates(webPath string) (*template.Template, error) {
	glob := filepath.Join(webPath, "template/*.gotmpl")

	return template.New("t").Funcs(template.FuncMap{
		"has":     templateHas,
		"getOr":   templateGetOr,
		"get":     templateGet,
		"content": templateContent,
	}).ParseGlob(glob)
}

func templateHas(m map[string][]interface{}, key string) bool {
	_, ok := get(m, key)

	return ok
}

func templateGetOr(m map[string][]interface{}, key string, or interface{}) interface{} {
	if value, ok := get(m, key); ok {
		return value
	}

	return or
}

func templateGet(m map[string][]interface{}, key string) interface{} {
	value, _ := get(m, key)

	return value
}

func templateContent(m map[string][]interface{}) interface{} {
	contents, ok := m["content"]

	if ok && len(contents) > 0 {
		if content, ok := contents[0].(string); ok {
			return content
		}

		if content, ok := contents[0].(map[string]interface{}); ok {
			if html, ok := content["html"]; ok {
				return template.HTML(html.(string))
			}

			if text, ok := content["text"]; ok {
				return text
			}
		}
	}

	return ""
}

func get(value interface{}, key string) (interface{}, bool) {
	if key == "" {
		return value, true
	}

	parts := strings.SplitN(key, ".", 2)

	if typed, ok := value.(map[string][]interface{}); ok {
		next, ok := typed[parts[0]]

		if len(next) == 0 {
			return nil, false
		}

		if ok && len(parts) == 2 {
			return get(next[0], parts[1])
		}

		return next[0], ok
	}

	if typed, ok := value.(map[string]interface{}); ok {
		next, ok := typed[parts[0]]

		if ok && len(parts) == 2 {
			return get(next, parts[1])
		}

		return next, ok
	}

	// if an array get the first value
	if typed, ok := value.([]interface{}); ok {
		if len(typed) > 0 {
			return get(typed[0], key)
		}

		return nil, false
	}

	return nil, false
}
