package blog

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/feeds"
	"hawx.me/code/numbersix"
	"hawx.me/code/route"
)

type Config struct {
	Me          string
	Name        string
	Title       string
	Description string
	BaseURL     *url.URL
	MediaURL    *url.URL
	DbPath      string
	MediaDir    string
	HubURL      string
}

type Blog struct {
	config        Config
	closer        io.Closer
	entries       *numbersix.DB
	mentions      *numbersix.DB
	syndicators   map[string]Syndicator
	templates     *template.Template
	citeResolvers []CiteResolver
	cardResolvers []CardResolver
	hubPublisher  HubPublisher
}

func New(
	config Config,
	db *sql.DB,
	templates *template.Template,
	hubPublisher HubPublisher,
	silos []interface{},
) (*Blog, error) {
	entries, err := numbersix.For(db, "entries")
	if err != nil {
		return nil, err
	}

	mentions, err := numbersix.For(db, "mentions")
	if err != nil {
		return nil, err
	}

	var (
		cardResolvers []CardResolver
		citeResolvers []CiteResolver
		syndicators   = map[string]Syndicator{}
	)

	for _, silo := range silos {
		if v, ok := silo.(CiteResolver); ok {
			citeResolvers = append(citeResolvers, v)
		}
		if v, ok := silo.(CardResolver); ok {
			cardResolvers = append(cardResolvers, v)
		}
		if v, ok := silo.(Syndicator); ok {
			syndicators[v.UID()] = v
		}
	}

	return &Blog{
		config:        config,
		closer:        db,
		entries:       entries,
		mentions:      mentions,
		templates:     templates,
		syndicators:   syndicators,
		citeResolvers: citeResolvers,
		cardResolvers: cardResolvers,
		hubPublisher:  hubPublisher,
	}, nil
}

func (b *Blog) Close() error {
	return b.closer.Close()
}

func (b *Blog) BaseURL() string {
	return b.config.BaseURL.String()
}

func (b *Blog) absoluteURL(p string) string {
	u, _ := url.Parse(p)

	return b.config.BaseURL.ResolveReference(u).String()
}

func (b *Blog) Handler() http.Handler {
	baseURL := b.config.BaseURL
	indexURL := b.absoluteURL("/")
	feedAtomURL := b.absoluteURL("/feed/atom")
	feedJsonfeedURL := b.absoluteURL("/feed/jsonfeed")
	feedRssURL := b.absoluteURL("/feed/rss")

	route.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.Before(before)
		if err != nil {
			log.Println("ERR get-all;", err)
			return
		}

		olderThan := ""
		if len(posts) == 25 {
			olderThan = posts[len(posts)-1].Properties["published"][0].(string)
		} else if len(posts) == 0 {
			olderThan = "NOMORE"
		}

		w.Header().Add("Link", `<`+indexURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)

		if err := b.templates.ExecuteTemplate(w, "page_list.gotmpl", pageListCtx{
			Title:        b.config.Title,
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
		}); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/kind/:kind", func(w http.ResponseWriter, r *http.Request) {
		vars := route.Vars(r)

		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.KindBefore(vars["kind"], before)
		if err != nil {
			log.Println("ERR get-all;", err)
			return
		}

		olderThan := ""
		if len(posts) == 25 {
			olderThan = posts[len(posts)-1].Properties["published"][0].(string)
		} else if len(posts) == 0 {
			olderThan = "NOMORE"
		}

		if err := b.templates.ExecuteTemplate(w, "page_list.gotmpl", pageListCtx{
			Title:        b.config.Title,
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
			Kind:         vars["kind"],
		}); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/category/:category", func(w http.ResponseWriter, r *http.Request) {
		vars := route.Vars(r)

		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.CategoryBefore(vars["category"], before)
		if err != nil {
			log.Println("ERR get-all;", err)
			return
		}

		olderThan := ""
		if len(posts) == 25 {
			olderThan = posts[len(posts)-1].Properties["published"][0].(string)
		} else if len(posts) == 0 {
			olderThan = "NOMORE"
		}

		if err := b.templates.ExecuteTemplate(w, "page_list.gotmpl", pageListCtx{
			Title:        b.config.Title,
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
			Kind:         "",
			Category:     vars["category"],
		}); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/entry/:id", func(w http.ResponseWriter, r *http.Request) {
		vars := route.Vars(r)

		entry, err := b.EntryByUID(vars["id"])
		if err != nil {
			log.Printf("ERR get-entry id=%s; %v\n", vars["id"], err)
			return
		}

		if deleted, ok := entry["hx-deleted"]; ok && len(deleted) > 0 {
			http.Error(w, "gone", http.StatusGone)
			return
		}

		mentions, err := b.MentionsForEntry(baseURL.ResolveReference(r.URL).String())
		if err != nil {
			log.Printf("ERR get-entry-mentions url=%s; %v\n", r.URL.Path, err)
			return
		}

		if err := b.templates.ExecuteTemplate(w, "page_post.gotmpl", struct {
			Posts    GroupedPosts
			Entry    map[string][]interface{}
			Mentions []numbersix.Group
		}{
			Entry: entry,
			Posts: GroupedPosts{
				Type: "entry",
				Meta: entry,
			},
			Mentions: mentions,
		}); err != nil {
			log.Printf("ERR get-entry-render url=%s; %v\n", r.URL.Path, err)
		}
	})

	route.HandleFunc("/likes/:ymd", func(w http.ResponseWriter, r *http.Request) {
		ymd := route.Vars(r)["ymd"]

		likes, err := b.LikesOn(ymd)
		if err != nil {
			log.Printf("ERR likes-on ymd=%s; %v\n", ymd, err)
			return
		}

		if err := b.templates.ExecuteTemplate(w, "page_day.gotmpl", struct {
			Title string
			Items []numbersix.Group
		}{
			Title: "likes for " + ymd,
			Items: likes,
		}); err != nil {
			log.Printf("ERR likes-on-render ymd=%s; %v\n", ymd, err)
		}
	})

	route.HandleFunc("/feed/rss", func(w http.ResponseWriter, r *http.Request) {
		f, err := b.feed()
		if err != nil {
			log.Println("ERR feed-rss;", err)
			return
		}

		rss, err := f.ToRss()
		if err != nil {
			log.Println("ERR feed-rss;", err)
			return
		}

		w.Header().Add("Link", `<`+feedRssURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)
		w.Header().Set("Content-Type", "application/rss+xml")
		io.WriteString(w, rss)
	})

	route.HandleFunc("/feed/atom", func(w http.ResponseWriter, r *http.Request) {
		f, err := b.feed()
		if err != nil {
			log.Println("ERR feed-atom;", err)
			return
		}

		atom, err := f.ToAtom()
		if err != nil {
			log.Println("ERR feed-atom;", err)
			return
		}

		w.Header().Add("Link", `<`+feedAtomURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)
		w.Header().Set("Content-Type", "application/atom+xml")
		io.WriteString(w, atom)
	})

	route.HandleFunc("/feed/jsonfeed", func(w http.ResponseWriter, r *http.Request) {
		f, err := b.feed()
		if err != nil {
			log.Println("ERR feed-jsonfeed;", err)
			return
		}

		json, err := f.ToJSON()
		if err != nil {
			log.Println("ERR feed-jsonfeed;", err)
			return
		}

		w.Header().Add("Link", `<`+feedJsonfeedURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, json)
	})

	// route.Handle("/:year/:month/:date/:slug")

	return route.Default
}

func (b *Blog) feed() (*feeds.Feed, error) {
	feed := &feeds.Feed{
		Title:   b.config.Title,
		Link:    &feeds.Link{Href: b.config.BaseURL.String()},
		Author:  &feeds.Author{Name: b.config.Name},
		Created: time.Now(),
	}

	posts, err := b.Before(time.Now().UTC())
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		relURL, _ := url.Parse(post.Properties["url"][0].(string))
		absURL := b.config.BaseURL.ResolveReference(relURL)

		createdAt, _ := time.Parse(time.RFC3339, post.Properties["published"][0].(string))

		feed.Items = append(feed.Items, &feeds.Item{
			Title:       templateTitle(post.Properties),
			Link:        &feeds.Link{Href: absURL.String()},
			Description: "",
			Created:     createdAt,
		})
	}

	return feed, nil
}

type pageListCtx struct {
	Title        string
	GroupedPosts []GroupedPosts
	OlderThan    string
	ShowLatest   bool
	Kind         string
	Category     string
}
