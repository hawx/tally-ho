package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func nav() lmth.Node {
	return Nav(lmth.Attr{},
		A(lmth.Attr{"class": "home", "href": "https://hawx.me/"}, lmth.Text("~hawx")),
		A(lmth.Attr{"href": "/"}, lmth.Text("blog")),
		A(lmth.Attr{"href": "https://hawx.me/code"}, lmth.Text("code")),
		A(lmth.Attr{"href": "https://hawx.me/talks"}, lmth.Text("talks")),
		A(lmth.Attr{"href": "https://ihkh.hawx.me/"}, lmth.Text("ihkh")),
		A(lmth.Attr{"href": "https://river.hawx.me/"}, lmth.Text("river")),
		A(lmth.Attr{"href": "https://garden.hawx.me/"}, lmth.Text("garden")),
		A(lmth.Attr{"href": "https://trobble.hawx.me/"}, lmth.Text("trobble")),
		A(lmth.Attr{"href": "https://hawx.me/media-diet"}, lmth.Text("media diet")),
	)
}

func buttons(goBack bool) lmth.Node {
	backLink := Span(lmth.Attr{})
	if goBack {
		backLink = A(lmth.Attr{"href": "/"}, lmth.Text("↑ Back to blog"))
	}

	return Div(lmth.Attr{"class": "buttons"},
		backLink,
		Div(lmth.Attr{},
			A(lmth.Attr{"href": "/likes"}, lmth.Text("likes")),
			A(lmth.Attr{"href": "/mentions"}, lmth.Text("mentions")),
		),
	)
}

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
