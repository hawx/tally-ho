package page

import (
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

func Mentions(data MentionsData) lmth.Node {
	var mentionsNode lmth.Node
	if data.OlderThan == "NOMORE" {
		mentionsNode = P(lmth.Attr{},
			lmth.Text("👏 You have reached the end. Try going back to the "),
			A(lmth.Attr{"class": "latest", "href": "/"},
				lmth.Text("Latest"),
			),
			lmth.Text("."),
		)
	} else {
		mentionsNode = Ul(lmth.Attr{"class": "mentions"},
			lmth.Map(func(item numbersix.Group) lmth.Node {
				name := " mentioned by "
				if mfutil.Has(item.Properties, "in-reply-to") {
					name = " reply from "
				} else if mfutil.Has(item.Properties, "repost-of") {
					name = " reposted by "
				} else if mfutil.Has(item.Properties, "like-of") {
					name = " liked by "
				}

				subject := item.Subject
				if mfutil.Has(item.Properties, "author") {
					if mfutil.Has(item.Properties, "author.properties.name") {
						subject = templateGet(item.Properties, "author.properties.name")
					} else {
						subject = templateGet(item.Properties, "author.properties.url")
					}
				}

				return Li(lmth.Attr{},
					A(lmth.Attr{"class": "target", "href": templateGet(item.Properties, "hx-target")},
						lmth.Text(templateGet(item.Properties, "hx-target")),
					),
					lmth.Text(name),
					A(lmth.Attr{"href": item.Subject},
						lmth.Text(subject),
					),
				)
			}, data.Items),
		)
	}

	return Html(lmth.Attr{"lang": "en"},
		pageHead(data.Title),
		Body(lmth.Attr{"class": "no-hero"},
			header(),
			P(lmth.Attr{"class": "page"},
				lmth.Text("mentions"),
			),
			Main(lmth.Attr{},
				mentionsNode,
			),
			Nav(lmth.Attr{"class": "arrows"},
				lmth.Toggle(data.OlderThan != "",
					A(lmth.Attr{"class": "older", "href": "?before=" + data.OlderThan},
						lmth.Text("Older"),
					),
				),
				lmth.Toggle(data.ShowLatest,
					A(lmth.Attr{"class": "latest", "href": "/mentions"},
						lmth.Text("Latest"),
					),
				),
			),
		),
		pageFooter(),
	)
}
