package page

type Context struct {
	basePath string

	// Name for the whole site, I'm using it like a nickname but it could be anything.
	Name string
	// Author is the name to be listed alongside entries as the publisher.
	Author string
	// Links are displayed in the header as a nice grid.
	Links []ContextLink
	// Copyright is shown in the footer.
	Copyright string
}

type ContextLink struct {
	Name string
	URL  string
}

func (c Context) WithPath(p string) Context {
	c.basePath = p
	return c
}

func (c Context) Path(p string) string {
	return c.basePath + p
}
