package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func nav(ctx Context) lmth.Node {
	links := make([]lmth.Node, len(ctx.Links)+1)
	links[0] = A(lmth.Attr{"class": "home", "href": "/"}, lmth.Text(ctx.Name))
	for i, link := range ctx.Links {
		links[i+1] = A(lmth.Attr{"href": link.URL}, lmth.Text(link.Name))
	}

	return Nav(lmth.Attr{}, links...)
}

func buttonsEmpty() lmth.Node {
	return Span(lmth.Attr{})
}

func buttonsBackToPosts() lmth.Node {
	return A(lmth.Attr{"href": "/posts"}, lmth.Text("↑ Back to posts"))
}

func buttonsLikesFor(formattedTime string) lmth.Node {
	return Span(lmth.Attr{"class": "page"},
		lmth.Text("likes for "),
		Strong(lmth.Attr{}, lmth.Text(formattedTime)),
	)
}

func buttons(left lmth.Node) lmth.Node {
	return Div(lmth.Attr{"class": "buttons"},
		left,
		Div(lmth.Attr{},
			A(lmth.Attr{"href": "/posts"}, lmth.Text("all")),
			A(lmth.Attr{"href": "/kind/like"}, lmth.Text("likes")),
			A(lmth.Attr{"href": "/mentions"}, lmth.Text("mentions")),
		),
	)
}
