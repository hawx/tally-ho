package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func nav(ctx Context) lmth.Node {
	links := make([]lmth.Node, len(ctx.Links)+1)
	links[0] = A(lmth.Attr{"class": "home", "href": ctx.URL}, lmth.Text(ctx.Name))
	for i, link := range ctx.Links {
		links[i+1] = A(lmth.Attr{"href": link.URL}, lmth.Text(link.Name))
	}

	return Nav(lmth.Attr{}, links...)
}

func buttons(goBack bool) lmth.Node {
	backLink := Span(lmth.Attr{})
	if goBack {
		backLink = A(lmth.Attr{"href": "/posts"}, lmth.Text("↑ Back to posts"))
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
