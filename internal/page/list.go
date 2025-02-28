package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
)

type ListData struct {
	Title        string
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

func List(data ListData) lmth.Node {
	var bodyNodes []lmth.Node
	var crumbs []string

	if data.Kind != "" {
		bodyNodes = append(bodyNodes, P(lmth.Attr{"class": "page"},
			lmth.Text("kind "),
			Strong(lmth.Attr{}, lmth.Text(data.Kind)),
		))
		crumbs = append(crumbs, "kind", "", data.Kind, "/kind/"+data.Kind)
	}
	if data.Category != "" {
		bodyNodes = append(bodyNodes, P(lmth.Attr{"class": "page"},
			lmth.Text("category "),
			Strong(lmth.Attr{}, lmth.Text(data.Category)),
		))
		crumbs = append(crumbs, "category", "", data.Category, "/category/"+data.Category)
	}

	if data.OlderThan == "NOMORE" {
		bodyNodes = append(bodyNodes, P(lmth.Attr{},
			lmth.Text("👏 You have reached the end. Try going back to the "),
			A(lmth.Attr{"class": "latest", "href": "/"}, lmth.Text("Latest")),
			lmth.Text("."),
		))
	} else {
		for _, grouping := range data.GroupedPosts {
			bodyNodes = append(bodyNodes, entryGrouping(grouping))
		}

		bodyNodes = append(bodyNodes, Nav(lmth.Attr{"class": "arrows"},
			lmth.Toggle(data.OlderThan != "",
				A(lmth.Attr{"class": "older", "href": "?before=" + data.OlderThan},
					lmth.Text("Older"),
				)),
			lmth.Toggle(data.ShowLatest, A(lmth.Attr{"class": "latest", "href": "/"},
				lmth.Text("Latest"),
			))))
	}

	return Html(lmth.Attr{"lang": "en"},
		pageHead(data.Title),
		Body(lmth.Attr{"class": "no-hero"},
			header(),
			Main(lmth.Attr{},
				bodyNodes...,
			),
		),
		pageFooter(crumbs...),
	)
}

func pageHead(title string, nodes ...lmth.Node) lmth.Node {
	def := []lmth.Node{
		Meta(lmth.Attr{"charset": "utf-8"}),
		Title(lmth.Attr{}, lmth.Text(title)),
		Meta(lmth.Attr{"content": "width=device-width, initial-scale=1", "name": "viewport"}),
		Link(lmth.Attr{"rel": "stylesheet", "href": "/public/styles.css", "type": "text/css"}),
		Link(lmth.Attr{"rel": "webmention", "href": "/-/webmention"}),
		Link(lmth.Attr{"rel": "alternative", "href": "/feed/atom", "type": "application/atom+xml"}),
		Link(lmth.Attr{"rel": "alternative", "href": "/feed/jsonfee", "type": "application/json"}),
		Link(lmth.Attr{"rel": "alternative", "href": "/feed/rss", "type": "application/rss+xml"}),
	}

	return Head(lmth.Attr{}, append(def, nodes...)...)
}

func pageFooter(crumbs ...string) lmth.Node {
	return Footer(lmth.Attr{},
		Nav(lmth.Attr{},
			Ul(lmth.Attr{},
				Li(lmth.Attr{},
					A(lmth.Attr{"href": "https://hawx.me/"}, lmth.Text("home")),
				),
				Li(lmth.Attr{},
					A(lmth.Attr{"href": "https://me.hawx.me/"}, lmth.Text("blog")),
				),
				lmth.Map2(func(i int, _ string) lmth.Node {
					if i%2 == 1 {
						return lmth.Text("")
					}

					if crumbs[i+1] == "" {
						return Li(lmth.Attr{}, lmth.Text(crumbs[i]))
					}

					return Li(lmth.Attr{},
						A(lmth.Attr{"href": crumbs[i+1]}, lmth.Text(crumbs[i])),
					)
				}, crumbs),
			),
		),
		P(lmth.Attr{"class": "copyright"}, lmth.Text("© 2024 Joshua Hawxwell.")),
	)
}
