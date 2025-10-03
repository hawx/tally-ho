package page

import (
	"log"

	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
	"hawx.me/code/numbersix"
	"hawx.me/code/tally-ho/internal/mfutil"
)

type MentionsData struct {
	Title      string
	Items      []numbersix.Group
	OlderThan  string
	ShowLatest bool
}

func Mentions(ctx Context, data MentionsData) lmth.Node {
	var bodyNodes lmth.Node

	if data.OlderThan == "NOMORE" {
		bodyNodes = P(lmth.Attr{},
			lmth.Text("👏 You have reached the end. Try going back to the "),
			A(lmth.Attr{"class": "latest", "href": "/mentions"},
				lmth.Text("Latest"),
			),
			lmth.Text("."),
		)
	} else {
		bodyNodes = lmth.Map(func(item numbersix.Group) lmth.Node {
			name := " mentioned "
			if mfutil.Has(item.Properties, "in-reply-to") {
				name = " replied to "
			} else if mfutil.Has(item.Properties, "repost-of") {
				name = " reposted "
			} else if mfutil.Has(item.Properties, "like-of") {
				name = " liked "
			}

			subject := item.Subject
			if mfutil.Has(item.Properties, "author") {
				if mfutil.Has(item.Properties, "author.properties.name") {
					subject = templateGet(item.Properties, "author.properties.name")
				} else {
					subject = templateGet(item.Properties, "author.properties.url")
				}
			}

			log.Println(item)

			return Article(lmth.Attr{"class": "mention"},
				H2(lmth.Attr{"class": "p-summary"},
					A(lmth.Attr{"href": item.Subject},
						lmth.Text(subject),
					),
					lmth.Text(name),
					A(lmth.Attr{"class": "target", "href": templateGet(item.Properties, "hx-target")},
						lmth.Text(templateGet(item.Properties, "hx-target")),
					),
				),
				entryMeta(item.Properties),
			)
		}, data.Items)
	}

	return Html(lmth.Attr{"lang": "en"},
		postsHead(data.Title),
		Body(lmth.Attr{},
			nav(ctx),
			buttons(Span(lmth.Attr{"class": "page"}, lmth.Text("mentions"))),
			Main(lmth.Attr{},
				bodyNodes,
			),
			Nav(lmth.Attr{"class": "buttons"},
				lmth.Toggle(data.OlderThan != "",
					A(lmth.Attr{"class": "older", "href": "?before=" + data.OlderThan},
						lmth.Text("← "),
						Span(lmth.Attr{}, lmth.Text("Older")),
					),
				),
				lmth.Toggle(data.ShowLatest,
					A(lmth.Attr{"class": "latest", "href": "/mentions"},
						Span(lmth.Attr{}, lmth.Text("Latest")),
						lmth.Text(" ⇥"),
					),
				),
			),
		),
		pageFooter(ctx),
	)
}
