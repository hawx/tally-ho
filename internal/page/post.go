package page

import (
	"fmt"
	"strings"
	"time"

	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
	"hawx.me/code/numbersix"
	"hawx.me/code/tally-ho/internal/mfutil"
)

type PostData struct {
	Posts    GroupedPosts
	Entry    map[string][]any
	Mentions []numbersix.Group
}

func Post(ctx Context, data PostData) lmth.Node {
	meta := data.Posts.Meta

	syndication := func() lmth.Node {
		syn, ok := meta["syndication"]
		if !ok || len(syn) == 0 {
			return lmth.Text("")
		}

		return Div(lmth.Attr{},
			lmth.Text("syndicated to "),
			lmth.Map2(func(i int, syndication any) lmth.Node {
				return lmth.Join(
					A(lmth.Attr{"class": "u-syndication", "href": syndication.(string)},
						lmth.Text(templateSyndicationName(syndication.(string))),
					),
					lmth.Toggle(i != len(syn)-1, lmth.Text(", ")),
				)
			}, syn),
		)
	}

	category := func() lmth.Node {
		cat, ok := meta["category"]
		if !ok || len(cat) == 0 {
			return lmth.Text("")
		}

		return Div(lmth.Attr{},
			lmth.Text("filed under "),
			lmth.Map2(func(i int, category any) lmth.Node {
				return lmth.Join(
					A(lmth.Attr{"class": "p-category", "href": ctx.Path("category/" + category.(string))},
						lmth.Text(category.(string)),
					),
					lmth.Toggle(i != len(cat)-1, lmth.Text(", ")),
				)
			}, cat),
		)
	}

	return Html(lmth.Attr{"lang": "en", "prefix": "og: http://ogp.me/ns#"},
		postsHead(ctx, templateTruncate(DecideTitle(data.Entry), 70),
			Meta(lmth.Attr{"property": "og:type", "content": "website"}),
			Meta(lmth.Attr{"property": "og:title", "content": DecideTitle(data.Entry)}),
			Meta(lmth.Attr{"property": "og:url", "content": templateGet(data.Entry, "url")}),
		),
		Body(lmth.Attr{},
			nav(ctx),
			buttons(buttonsBackToPosts(ctx)),
			Main(lmth.Attr{},
				Article(lmth.Attr{"class": "h-entry " + templateGet(meta, "hx-kind")},
					lmth.Join(
						entry(data.Posts.Meta),
						Div(lmth.Attr{"class": "expanded meta"},
							Div(lmth.Attr{},
								A(lmth.Attr{"href": ctx.Path("kind/" + templateGet(meta, "hx-kind"))},
									lmth.Text(templateGet(meta, "hx-kind")),
								),
								lmth.Text(" "),
								publishedUpdated(meta),
							),
							Div(lmth.Attr{},
								lmth.Text("by "),
								A(lmth.Attr{"class": "u-author h-card", "href": templateGet(meta, "author.properties.url")},
									lmth.Text(templateGet(meta, "author.properties.name")),
								),
							),
							lmth.Toggle(mfutil.Has(meta, "hx-client-id"),
								Div(lmth.Attr{},
									lmth.Text("using "),
									A(lmth.Attr{"href": templateGet(meta, "hx-client-id")},
										lmth.Text(templateGet(meta, "hx-client-id")),
									),
								)),
							syndication(),
							category(),
						),
						Details(lmth.Attr{"class": "meta"},
							Summary(lmth.Attr{},
								lmth.Text(fmt.Sprintf("Interactions (%d)", len(data.Mentions))),
							),
							Ol(lmth.Attr{},
								lmth.Map(func(mention numbersix.Group) lmth.Node {
									name := "mentioned by "
									if mfutil.Has(mention.Properties, "in-reply-to") {
										name = "reply from "
									} else if mfutil.Has(mention.Properties, "repost-of") {
										name = "reposted by "
									} else if mfutil.Has(mention.Properties, "like-of") {
										name = "liked by "
									}

									subject := mention.Subject
									if mfutil.Has(mention.Properties, "author") {
										if mfutil.Has(mention.Properties, "author.properties.name") {
											subject = templateGet(mention.Properties, "author.properties.name")
										} else {
											subject = templateGet(mention.Properties, "author.properties.url")
										}
									}

									return Li(lmth.Attr{},
										lmth.Text(name),
										A(lmth.Attr{"href": mention.Subject},
											lmth.Text(subject),
										),
									)
								}, data.Mentions),
							),
						),
					),
				),
			),
		),
		pageFooter(ctx),
	)
}

func publishedUpdated(meta map[string][]any) lmth.Node {
	publishedNode := lmth.Join(
		lmth.Text("published "),
		A(lmth.Attr{"class": "u-url", "href": templateGet(meta, "url"), "title": templateGet(meta, "published")},
			Time(lmth.Attr{"class": "dt-published", "datetime": templateGet(meta, "published")},
				lmth.Text(templateHumanDateTime(meta, "published")),
			),
		),
	)

	if !mfutil.Has(meta, "updated") {
		return publishedNode
	}

	return lmth.Join(
		Del(lmth.Attr{}, publishedNode),
		lmth.Text("updated "),
		Time(lmth.Attr{"class": "dt-updated", "datetime": templateGet(meta, "updated")},
			lmth.Text(templateHumanDateTime(meta, "updated")),
		),
	)
}

func templateTruncate(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length] + "…"
}

func templateHumanDateTime(m map[string][]any, key string) string {
	s, ok := mfutil.Get(m, key).(string)
	if !ok {
		return ""
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}

	return t.Format("January 02, 2006 at 15:04")
}

func templateSyndicationName(u string) string {
	if strings.HasPrefix(u, "https://www.flickr.com/") {
		return "Flickr"
	}

	if strings.HasPrefix(u, "https://github.com/") {
		return "GitHub"
	}

	return u
}
