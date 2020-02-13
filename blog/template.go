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
		"humanDateTime":   templateHumanDateTime,
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
	if mfutil.Has(m, "content.html") {
		return template.HTML(mfutil.Get(m, "content.html").(string))
	}

	return mfutil.Get(m, "content.text", "content").(string)
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

func templateHumanDateTime(m map[string][]interface{}, key string) string {
	s, ok := mfutil.Get(m, key).(string)
	if !ok {
		return ""
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("January 02, 2006 at 15:04")
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
	prefix := ""
	defalt := "a post"

	switch mfutil.Get(m, "hx-kind").(string) {
	case "rsvp":
		return templateHumanRSVP(m) + " to " + templateGetOr(m, "name", "an event").(string)
	case "reply":
		return "replied to " + mfutil.Get(m,
			"in-reply-to.properties.name",
			"in-reply-to.properties.url",
			"in-reply-to").(string)
	case "repost":
		return "reposted " + mfutil.Get(m,
			"repost-of.properties.name",
			"repost-of.properties.url",
			"repost-of").(string)
	case "like":
		return "liked " + mfutil.Get(m,
			"like-of.properties.name",
			"like-of.properties.url",
			"like-of").(string)
	case "bookmark":
		return "bookmarked " + mfutil.Get(m,
			"bookmark-of.properties.name",
			"bookmark-of.properties.url",
			"bookmark-of").(string)
	case "video":
		prefix = "video: "
		defalt = "a video"
	case "photo":
		prefix = "photo: "
		defalt = "a photo"
	case "read":
		return templateHumanReadStatus(m) + " " +
			mfutil.Get(m, "read-of.properties.name").(string) + " by " +
			mfutil.Get(m, "read-of.properties.author").(string)
	case "drank":
		return "drank " + mfutil.Get(m, "drank.properties.name").(string)
	case "checkin":
		return "checked in to " + mfutil.Get(m, "checkin.properties.name").(string)
	}

	if mfutil.Has(m, "name") {
		return prefix + mfutil.Get(m, "name").(string)
	}

	if content, ok := mfutil.Get(m, "content.text", "content").(string); ok {
		return prefix + templateTruncate(content, 140)
	}

	return defalt
}

func templateTruncate(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length] + "…"
}
