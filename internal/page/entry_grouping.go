package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func entryGrouping(grouping GroupedPosts) lmth.Node {
	if grouping.Type == "like" {
		likedPosts := []lmth.Node{lmth.Text("liked ")}

		for _, post := range grouping.Posts {
			name := templateGet(post, "like-of.properties.name")
			if name == "" {
				name = templateGet(post, "like-of.properties.url")
			}

			likedPosts = append(likedPosts, Span(lmth.Attr{"class": "h-entry"},
				Span(lmth.Attr{"class": "hidden"}, lmth.Text("liked ")),
				A(lmth.Attr{"class": "u-like-of", "href": templateGet(post, "like-of.properties.url")},
					lmth.Text(name),
				),
				lmth.Text(" "),
				A(lmth.Attr{"class": "u-url", "href": templateGet(post, "url")},
					lmth.Text("at "),
					Time(lmth.Attr{"class": "dt-published", "datetime": templateGet(post, "published")},
						lmth.Text(formatTime(templateGet(post, "published"))),
					),
				),
				A(lmth.Attr{"class": "u-author h-card hidden", "href": templateGet(post, "author.properties.url")},
					lmth.Text(templateGet(post, "author.properties.name")),
				),
			))
		}

		return Article(lmth.Attr{"class": "like"},
			H2(lmth.Attr{},
				likedPosts...,
			),
			Div(lmth.Attr{"class": "meta right"},
				A(lmth.Attr{"href": templateGet(grouping.Meta, "url"), "title": templateGet(grouping.Meta, "published")},
					Time(lmth.Attr{"datetime": templateGet(grouping.Meta, "published")},
						lmth.Text(formatHumanDate(templateGet(grouping.Meta, "published"))),
					),
				),
			),
		)
	} else {
		return Article(lmth.Attr{"class": "h-entry " + templateGet(grouping.Meta, "hx-kind")},
			lmth.Join(
				entry(grouping.Meta),
				entryMeta(grouping.Meta),
			),
		)
	}
}
