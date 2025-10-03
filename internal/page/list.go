package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

type ListData struct {
	GroupedPosts []GroupedPosts
	OlderThan    string
	ShowLatest   bool
	Kind         string
	Category     string
}

type GroupedPosts struct {
	Type  string
	Posts []map[string][]any
	Meta  map[string][]any
}

func List(ctx Context, data ListData) lmth.Node {
	var bodyNodes []lmth.Node

	buttonsLeft := buttonsEmpty()
	if data.Kind != "" {
		buttonsLeft = Span(lmth.Attr{"class": "page"},
			lmth.Text("kind "),
			Strong(lmth.Attr{}, lmth.Text(data.Kind)),
		)
	}
	if data.Category != "" {
		buttonsLeft = Span(lmth.Attr{"class": "page"},
			lmth.Text("category "),
			Strong(lmth.Attr{}, lmth.Text(data.Category)),
		)
	}

	var bottomButtons lmth.Node
	if data.OlderThan == "NOMORE" {
		bodyNodes = append(bodyNodes, P(lmth.Attr{},
			lmth.Text("👏 You have reached the end. Try going back to the "),
			A(lmth.Attr{"class": "latest", "href": "/posts"}, lmth.Text("Latest")),
			lmth.Text("."),
		))
	} else {
		for _, grouping := range data.GroupedPosts {
			bodyNodes = append(bodyNodes, entryGrouping(grouping))
		}

		bottomButtons = Div(lmth.Attr{"class": "buttons"},
			lmth.Toggle(data.OlderThan != "",
				A(lmth.Attr{"class": "older", "href": "?before=" + data.OlderThan},
					lmth.Text("← "),
					Span(lmth.Attr{}, lmth.Text("Older")),
				)),
			lmth.Toggle(data.ShowLatest, A(lmth.Attr{"class": "latest", "href": "/posts"},
				Span(lmth.Attr{}, lmth.Text("Latest")),
				lmth.Text(" ⇥"),
			)))
	}

	return Html(lmth.Attr{"lang": "en"},
		postsHead(ctx.Name+" posts"),
		Body(lmth.Attr{},
			nav(ctx),
			buttons(buttonsLeft),
			Main(lmth.Attr{},
				bodyNodes...,
			),
			bottomButtons,
		),
		pageFooter(ctx),
	)
}

func postsHead(title string, nodes ...lmth.Node) lmth.Node {
	def := []lmth.Node{
		Link(lmth.Attr{"rel": "webmention", "href": "/-/webmention"}),
		Link(lmth.Attr{"rel": "alternative", "href": "/feed/atom", "type": "application/atom+xml"}),
		Link(lmth.Attr{"rel": "alternative", "href": "/feed/jsonfee", "type": "application/json"}),
		Link(lmth.Attr{"rel": "alternative", "href": "/feed/rss", "type": "application/rss+xml"}),
	}

	return pageHead(title, append(def, nodes...)...)
}

func pageHead(title string, nodes ...lmth.Node) lmth.Node {
	def := []lmth.Node{
		Meta(lmth.Attr{"charset": "utf-8"}),
		Title(lmth.Attr{}, lmth.Text(title)),
		Meta(lmth.Attr{"content": "width=device-width, initial-scale=1", "name": "viewport"}),
		Link(lmth.Attr{"rel": "stylesheet", "href": "/public/styles.css", "type": "text/css"}),
	}

	return Head(lmth.Attr{}, append(def, nodes...)...)
}

func pageFooter(ctx Context) lmth.Node {
	return Footer(lmth.Attr{},
		Span(lmth.Attr{"class": "copyright"}, lmth.Text(ctx.Copyright)),
	)
}
