package page

import (
	"time"

	"hawx.me/code/tally-ho/internal/mfutil"
)

func conv[T any](x any) T {
	v, _ := x.(T)
	return v
}

func DecideTitle(m map[string][]any) string {
	prefix := ""
	defalt := "a post"

	switch mfutil.Get(m, "hx-kind").(string) {
	case "rsvp":
		return formatHumanRSVP(templateGet(m, "rsvp")) + " to " + templateGetOr(m, "name", "an event")
	case "reply":
		return "replied to " + conv[string](mfutil.Get(m,
			"in-reply-to.properties.name",
			"in-reply-to.properties.url",
			"in-reply-to"))
	case "repost":
		return "reposted " + conv[string](mfutil.Get(m,
			"repost-of.properties.name",
			"repost-of.properties.url",
			"repost-of"))
	case "like":
		return "liked " + conv[string](mfutil.Get(m,
			"like-of.properties.name",
			"like-of.properties.url",
			"like-of"))
	case "bookmark":
		return "bookmarked " + conv[string](mfutil.Get(m,
			"bookmark-of.properties.name",
			"bookmark-of.properties.url",
			"bookmark-of"))
	case "video":
		prefix = "video: "
		defalt = "a video"
	case "photo":
		prefix = "photo: "
		defalt = "a photo"
	case "read":
		if mfutil.Has(m, "read-of.properties.author") {
			return formatReadStatus(templateGet(m, "read-status")) + " " +
				conv[string](mfutil.Get(m, "read-of.properties.name")) + " by " +
				conv[string](mfutil.Get(m, "read-of.properties.author"))
		}
		return formatReadStatus(templateGet(m, "read-status")) + " " +
			conv[string](mfutil.Get(m, "read-of.properties.name"))
	case "drank":
		return "drank " + conv[string](mfutil.Get(m, "drank.properties.name"))
	case "checkin":
		return "checked in to " + conv[string](mfutil.Get(m, "checkin.properties.name"))
	}

	if name, ok := mfutil.Get(m, "name", "content.text", "content").(string); ok {
		return prefix + name
	}

	return defalt
}

func formatHumanDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("January 02, 2006")
}

func formatTime(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("15:04")
}

func formatHumanRSVP(s string) string {
	switch s {
	case "yes":
		return "going"
	case "no":
		return "not going"
	default:
		return "might be going"
	}
}

func formatReadStatus(s string) string {
	switch s {
	case "to-read":
		return "want to read"
	case "reading":
		return "reading"
	default:
		return "read"
	}
}
