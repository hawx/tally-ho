package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

func nav(ctx Context) lmth.Node {
	links := make([]lmth.Node, len(ctx.Links)+1)
	links[0] = Span(lmth.Attr{"class": "home"},
		A(lmth.Attr{"href": ctx.Path("")}, lmth.Text(ctx.Name)))
	for i, link := range ctx.Links {
		links[i+1] = Span(lmth.Attr{},
			A(lmth.Attr{"href": link.URL}, lmth.Text(link.Name)))
	}

	return Nav(lmth.Attr{}, links...)
}

func buttonsEmpty() lmth.Node {
	return Span(lmth.Attr{})
}

func buttonsBackToPosts(ctx Context) lmth.Node {
	return A(lmth.Attr{"href": ctx.Path("")}, lmth.Text("↑ Back to posts"))
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
	)
}
