package page

import (
	"time"

	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
	"hawx.me/code/numbersix"
)

type DayData struct {
	Ymd   string
	Items []numbersix.Group
}

func Day(ctx Context, data DayData) lmth.Node {
	formattedTime := data.Ymd

	t, err := time.Parse(time.DateOnly, data.Ymd)
	if err == nil {
		formattedTime = t.Format("January 02, 2006")
	}

	return Html(lmth.Attr{"lang": "en"},
		postsHead("likes for "+formattedTime),
		Body(lmth.Attr{},
			nav(ctx),
			buttons(buttonsLikesFor(formattedTime)),
			Main(lmth.Attr{},
				lmth.Map(func(group numbersix.Group) lmth.Node {
					return Article(lmth.Attr{"class": "h-entry " + templateGet(group.Properties, "hx-kind")},
						entry(group.Properties),
						entryMeta(group.Properties),
					)
				}, data.Items),
			),
		),
		pageFooter(ctx),
	)
}
