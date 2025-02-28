package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func header() lmth.Node {
	return Header(lmth.Attr{"class": "h-card full-width"},
		H1(lmth.Attr{"class": "p-name"},
			A(lmth.Attr{"class": "u-url u-uid", "href": "https://hawx.me/"},
				Span(lmth.Attr{"class": "p-given-name"}, lmth.Text("Joshua")),
				lmth.Text(" "),
				Span(lmth.Attr{"class": "p-family-name"}, lmth.Text("Hawxwell")),
			),
		),
		Span(lmth.Attr{"class": "app-hidden p-nickname"}, lmth.Text("hawx")),
		Img(lmth.Attr{"class": "app-hidden u-photo", "src": "/avatar.jpg"}),
		Ul(lmth.Attr{},
			Li(lmth.Attr{},
				A(lmth.Attr{"href": "#"}, lmth.Text("mentions")),
			),
			Li(lmth.Attr{},
				A(lmth.Attr{"href": "#"}, lmth.Text("likes")),
			),
			Li(lmth.Attr{},
				A(lmth.Attr{"href": "#"}, lmth.Text("archive")),
			),
		),
	)
}
