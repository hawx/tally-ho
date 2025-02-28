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

func Day(data DayData) lmth.Node {
	formattedTime := data.Ymd

	t, err := time.Parse(time.DateOnly, data.Ymd)
	if err == nil {
		formattedTime = t.Format("January 02, 2006")
	}

	return Html(lmth.Attr{"lang": "en"},
		pageHead("likes for "+formattedTime),
		Body(lmth.Attr{"class": "no-hero"},
			header(),
			Main(lmth.Attr{},
				P(lmth.Attr{"class": "page"},
					lmth.Text("likes for "),
					Strong(lmth.Attr{}, lmth.Text(formattedTime)),
				),
				lmth.Map(func(group numbersix.Group) lmth.Node {
					return Article(lmth.Attr{"class": "h-entry " + templateGet(group.Properties, "hx-kind")},
						entry(group.Properties),
						entryMeta(group.Properties),
					)
				}, data.Items),
			),
		),
		pageFooter("likes", "", formattedTime, "/likes/"+data.Ymd),
	)
}
