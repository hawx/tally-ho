package blog

import (
	"html/template"
	"path/filepath"
	"strings"
	"time"

	"hawx.me/code/tally-ho/internal/mfutil"
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
		"humanReadStatus": templateHumanReadStatus,
		"time":            templateTime,
		"syndicationName": templateSyndicationName,
		"withEnd":         templateWithEnd,
		"title":           templateTitle,
		"truncate":        templateTruncate,
	}).ParseGlob(glob)
}

func templateHas(v interface{}, key string) bool {
	m, ok := v.(map[string][]interface{})
	if !ok {
		return false
	}

	return mfutil.Has(m, key)
}

func templateGetOr(m map[string][]interface{}, key string, or interface{}) interface{} {
	if value, ok := mfutil.SafeGet(m, key); ok {
		return value
	}

	return or
}

func templateGet(m map[string][]interface{}, key string) interface{} {
	return mfutil.Get(m, key)
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

func templateHumanDate(m map[string][]interface{}, key string) string {
	s, ok := mfutil.Get(m, key).(string)
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
	s, ok := mfutil.Get(m, "rsvp").(string)
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

func templateHumanReadStatus(m map[string][]interface{}) string {
	s, ok := mfutil.Get(m, "read-status").(string)
	if !ok {
		return ""
	}

	switch s {
	case "to-read":
		return "want to read"
	case "reading":
		return "reading"
	default:
		return "read"
	}
}

func templateTime(m map[string][]interface{}, key string) string {
	s, ok := mfutil.Get(m, key).(string)
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
	switch templateGet(m, "hx-kind").(string) {
	case "like":
		if templateHas(m, "like-of.properties.name") {
			return "liked " + templateGet(m, "like-of.properties.name").(string)
		} else {
			return "liked " + templateGet(m, "like-of.properties.url").(string)
		}
	case "rsvp":
		return templateHumanRSVP(m) + " to " + templateGetOr(m, "name", "an event").(string)
	case "read":
		return templateHumanReadStatus(m) + " " + templateGet(m, "read-of.properties.name").(string) + " by " + templateGet(m, "read-of.properties.author").(string)
	case "drank":
		return "drank " + templateGet(m, "drank.properties.name").(string)
	case "checkin":
		return "checked in to " + templateGet(m, "checkin.properties.name").(string)
	}

	if templateHas(m, "name") {
		return templateGet(m, "name").(string)
	}

	if templateHas(m, "content") {
		if content, ok := templateContent(m).(string); ok {
			return templateTruncate(content, 140)
		}
	}

	return "a post"
}

func templateTruncate(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length] + "…"
}
