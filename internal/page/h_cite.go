package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
	"hawx.me/code/tally-ho/internal/mfutil"
)

func hCite(meta map[string]any) lmth.Node {
	if !mfutil.Has(meta, "content") && !mfutil.Has(meta, "photo") {
		return lmth.Text("")
	}

	return Div(lmth.Attr{"class": "h-cite"},
		lmth.Toggle(len(mfutil.GetAll(meta, "photo")) == 1,
			Img(lmth.Attr{"src": templateGet(meta, "photo")}),
		),
		lmth.Toggle(mfutil.Has(meta, "author.properties.name"),
			P(lmth.Attr{"class": "p-author h-card"},
				A(lmth.Attr{"class": "p-name u-url", "href": templateGet(meta, "author.properties.url")},
					lmth.Text(templateGet(meta, "author.properties.name")),
				),
				lmth.Text(":"),
			),
		),
		lmth.Toggle(len(mfutil.GetAll(meta, "photo")) > 1,
			lmth.Map(func(s any) lmth.Node {
				return Img(lmth.Attr{"src": s.(string)})
			}, mfutil.GetAll(meta, "photo")),
		),
		lmth.Toggle(mfutil.Has(meta, "content"),
			Div(lmth.Attr{"class": "e-content"},
				templateContent(meta),
			),
		),
		lmth.Toggle(mfutil.Has(meta, "published"),
			Div(lmth.Attr{"class": "meta"},
				A(lmth.Attr{"class": "u-url", "href": templateGet(meta, "url")},
					Time(lmth.Attr{"class": "dt-published", "datetime": templateGet(meta, "published")},
						lmth.Text(formatHumanDate(templateGet(meta, "published"))),
					),
				),
			),
		),
	)
}
