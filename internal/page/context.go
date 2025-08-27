package page

type Context struct {
	Name      string
	URL       string
	Links     []ContextLink
	Copyright string
	About     string

	BlogTitle string
}

type ContextLink struct {
	Name string
	URL  string
}
