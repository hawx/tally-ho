package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
	"hawx.me/code/tally-ho/internal/mfutil"
)

func entry(meta map[string][]any) lmth.Node {
	var nodes []lmth.Node

	if mfutil.Has(meta, "rsvp") {
		nodes = append(nodes, H2(lmth.Attr{"class": "p-summary"},
			Data(lmth.Attr{"class": "p-rsvp", "value": templateGet(meta, "rsvp")},
				lmth.Text(formatHumanRSVP(templateGet(meta, "rsvp"))),
			),
			lmth.Text(" to "),
			A(lmth.Attr{"class": "u-in-reply-to", "href": templateGet(meta, "in-reply-to")},
				lmth.Text(templateGetOr(meta, "name", "an event")),
			),
		))
	} else if mfutil.Has(meta, "like-of") {
		nodes = append(nodes, entryH2(meta, "liked", "like-of"))
	} else if mfutil.Has(meta, "bookmark-of") {
		nodes = append(nodes, entryH2(meta, "bookmarked", "bookmark-of"))
	} else if mfutil.Has(meta, "in-reply-to") {
		nodes = append(nodes, entryH2(meta, "replied to", "in-reply-to"))
	} else if mfutil.Has(meta, "repost-of") {
		nodes = append(nodes, entryH2(meta, "reposted", "repost-of"))
	} else if mfutil.Has(meta, "read-of") {
		citeNodes := []lmth.Node{
			Strong(lmth.Attr{"class": "p-name"},
				lmth.Text(templateGet(meta, "read-of.properties.name")),
			),
		}

		if mfutil.Has(meta, "read-of.properties.author") {
			citeNodes = append(citeNodes,
				lmth.Text(" by "),
				Strong(lmth.Attr{"class": "p-author"},
					lmth.Text(templateGet(meta, "read-of.properties.author")),
				),
			)
		}

		nodes = append(nodes, H2(lmth.Attr{"class": "p-summary"},
			lmth.Text(formatReadStatus(templateGet(meta, "read-status"))+" "),
			Span(lmth.Attr{"class": "h-cite"},
				citeNodes...,
			),
		))
	} else if mfutil.Has(meta, "drank") {
		nodes = append(nodes, H2(lmth.Attr{"class": "p-summary"},
			lmth.Text("drank "),
			Strong(lmth.Attr{},
				lmth.Text(templateGet(meta, "drank.properties.name")),
			),
		))
	} else if mfutil.Has(meta, "ate") {
		nodes = append(nodes, H2(lmth.Attr{"class": "p-summary"},
			lmth.Text("ate "),
			Strong(lmth.Attr{},
				lmth.Text(templateGet(meta, "ate.properties.name")),
			),
		))
	} else if mfutil.Has(meta, "checkin") {
		nodes = append(nodes, P(lmth.Attr{"class": "h-card p-summary"},
			H2(lmth.Attr{},
				lmth.Text("checked in to "),
				A(lmth.Attr{"class": "u-url p-name", "href": templateGet(meta, "checkin.properties.url")},
					lmth.Text(templateGet(meta, "checkin.properties.name")),
				),
				lmth.Text(" "),
				Span(lmth.Attr{"class": "full-address"},
					Span(lmth.Attr{"class": "p-street-address"},
						lmth.Text(templateGet(meta, "checking.properties.street-address")),
					),
					lmth.Text(", "),
					Span(lmth.Attr{"class": "p-locality"},
						lmth.Text(templateGet(meta, "checking.properties.locality")),
					),
					lmth.Text(", "),
					Span(lmth.Attr{"class": "p-country-name"},
						lmth.Text(templateGet(meta, "checking.properties.country-name")),
					),
				),
			),
		))
	}

	nodes = append(nodes, hCite(templateCite(meta)))

	if mfutil.Has(meta, "name") {
		nodes = append(nodes, H2(lmth.Attr{"class": "p-name"},
			A(lmth.Attr{"href": templateGet(meta, "url")},
				lmth.Text(templateGet(meta, "name")),
			),
		))
	}

	for _, photo := range meta["photo"] {
		if mfutil.Has(photo, "value") {
			nodes = append(nodes, Img(lmth.Attr{"src": templateGet(photo, "value"), "alt": templateGet(photo, "alt")}))
		} else {
			nodes = append(nodes, Img(lmth.Attr{"src": photo.(string)}))
		}
	}

	if mfutil.Has(meta, "content") {
		class := "e-content"
		if templateGet(meta, "hx-kind") == "note" {
			class += " p-name"
		}

		nodes = append(nodes, Div(lmth.Attr{"class": class},
			templateContent(meta),
		))
	}

	return lmth.Join(nodes...)
}

func entryH2(meta map[string][]any, intro string, key string) lmth.Node {
	return H2(lmth.Attr{"class": "p-summary"},
		lmth.Text(intro+" "),
		A(lmth.Attr{"class": "u-" + key, "href": templateGet(meta, key+".properties.url")},
			lmth.Text(templateGetOr(meta, key+".properties.name", templateGet(meta, key+".properties.url"))),
		),
	)
}

func templateGet(m any, key string) string {
	return conv[string](mfutil.Get(m, key))
}

func templateGetOr(m map[string][]any, key string, or string) string {
	if value, ok := mfutil.SafeGet(m, key); ok {
		return conv[string](value)
	}

	return or
}

func templateContent(m any) lmth.Node {
	if mfutil.Has(m, "content.html") {
		return lmth.RawText(conv[string](mfutil.Get(m, "content.html")))
	}

	if s, ok := mfutil.Get(m, "content.text", "content").(string); ok {
		return lmth.Text(s)
	}

	return lmth.Text("")
}

func templateCite(m map[string][]any) map[string]any {
	for _, value := range m {
		if t, ok := mfutil.Get(value, "type").(string); ok && t == "h-cite" {
			return conv[map[string]any](mfutil.Get(value, "properties"))
		}
	}

	return nil
}
