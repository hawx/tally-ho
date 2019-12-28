package blog

import (
	"html/template"
	"path/filepath"
	"strings"
	"time"
)

func ParseTemplates(webPath string) (*template.Template, error) {
	glob := filepath.Join(webPath, "template/*.gotmpl")

	return template.New("t").Funcs(template.FuncMap{
		"has":             templateHas,
		"getOr":           templateGetOr,
		"get":             templateGet,
		"content":         templateContent,
		"humanDate":       templateHumanDate,
		"humanRSVP":       templateHumanRSVP,
		"time":            templateTime,
		"syndicationName": templateSyndicationName,
		"withEnd":         templateWithEnd,
		"title":           templateTitle,
		"truncate":        templateTruncate,
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
	// if an array get the first value
	if typed, ok := value.([]interface{}); ok {
		if len(typed) > 0 {
			return get(typed[0], key)
		}

		return nil, false
	}

	// if no key then this must be what we were looking for
	if key == "" {
		return value, true
	}

	parts := strings.SplitN(key, ".", 2)

	if typed, ok := value.(map[string][]interface{}); ok {
		next, ok := typed[parts[0]]

		if !ok || len(next) == 0 {
			return nil, false
		}

		if len(parts) == 2 {
			return get(next[0], parts[1])
		}

		return get(next[0], "")
	}

	if typed, ok := value.(map[string]interface{}); ok {
		next, ok := typed[parts[0]]

		if !ok {
			return nil, ok
		}

		if len(parts) == 2 {
			return get(next, parts[1])
		}

		return get(next, "")
	}

	return nil, false
}

func templateHumanDate(m map[string][]interface{}, key string) string {
	v, _ := get(m, key)
	s, ok := v.(string)

	if !ok {
		return ""
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("January 02, 2006")
}

func templateHumanRSVP(m map[string][]interface{}) string {
	v, _ := get(m, "rsvp")
	s, ok := v.(string)

	if !ok {
		return ""
	}

	switch s {
	case "yes":
		return "going"
	case "no":
		return "not going"
	default:
		return "might be going"
	}
}

func templateTime(m map[string][]interface{}, key string) string {
	v, _ := get(m, key)
	s, ok := v.(string)

	if !ok {
		return ""
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("15:04")
}

func templateSyndicationName(u string) string {
	if strings.HasPrefix(u, "https://twitter.com/") {
		return "Twitter"
	}

	return u
}

type endEl struct {
	El  interface{}
	End bool
}

func templateWithEnd(l []interface{}) []endEl {
	r := make([]endEl, len(l))

	for i, e := range l {
		r[i] = endEl{El: e, End: i == len(l)-1}
	}

	return r
}

func templateTitle(m map[string][]interface{}) string {
	wrap := func(s string) string {
		if templateHas(m, "like-of") {
			return "Liked: " + s
		}

		return s
	}

	if templateHas(m, "name") {
		return wrap(templateGet(m, "name").(string))
	}

	if templateHas(m, "content") {
		if content, ok := templateContent(m).(string); ok {
			return wrap(templateTruncate(content, 140))
		}
	}

	return wrap("a post")
}

func templateTruncate(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length] + "…"
}
