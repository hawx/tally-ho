package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func entryMeta(meta map[string][]any) lmth.Node {
	return Div(lmth.Attr{"class": "meta right"},
		A(lmth.Attr{"class": "u-url", "href": templateGet(meta, "url"), "title": templateGet(meta, "published")},
			Time(lmth.Attr{"class": "dt-published", "datetime": templateGet(meta, "published")},
				lmth.Text(formatHumanDate(templateGet(meta, "published"))),
			),
		),
		A(lmth.Attr{"class": "u-author h-card hidden", "href": templateGet(meta, "author.properties.url")},
			lmth.Text(templateGet(meta, "author.properties.name")),
		),
	)
}
