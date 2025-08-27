package page

import (
	"hawx.me/code/lmth"
	. "hawx.me/code/lmth/elements"
	"hawx.me/code/tally-ho/internal/mfutil"
)

func HxPage(ctx Context, entry map[string][]any) lmth.Node {
	name, ok := mfutil.Get(entry, "name").(string)
	if !ok {
		return lmth.Node{}
	}

	content, ok := mfutil.Get(entry, "content.html").(string)
	if !ok {
		return lmth.Node{}
	}

	var hero lmth.Node
	if heroBackground, ok := mfutil.Get(entry, "hx-hero-background").(string); ok {
		if heroImg, ok := mfutil.Get(entry, "hx-hero-img").(string); ok {
			hero = Div(lmth.Attr{"class": "hero", "style": "background:" + heroBackground},
				Img(lmth.Attr{"src": heroImg}),
			)
		}
	}

	return Html(lmth.Attr{"lang": "en"},
		pageHead(name),
		Body(lmth.Attr{},
			nav(ctx),
			hero,
			Main(lmth.Attr{},
				lmth.RawText(content),
			),
		),
		pageFooter(ctx),
	)
}
