package blog

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/feeds"
	"hawx.me/code/numbersix"
	"hawx.me/code/route"
	"hawx.me/code/tally-ho/internal/page"
)

type Config struct {
	Me       string
	BaseURL  *url.URL
	MediaURL *url.URL
	HubURL   string
}

type Blog struct {
	local         bool
	config        Config
	pageCtx       page.Context
	closer        io.Closer
	entries       *numbersix.DB
	mentions      *numbersix.DB
	syndicators   map[string]Syndicator
	citeResolvers []CiteResolver
	cardResolvers []CardResolver
	hubPublisher  HubPublisher
}

func New(
	logger *slog.Logger,
	config Config,
	pageCtx page.Context,
	db *sql.DB,
	hubPublisher HubPublisher,
	silos []any,
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

	local := config.BaseURL.Hostname() == "localhost"
	if local {
		logger.Info("running in local mode")
	}

	return &Blog{
		local:         local,
		config:        config,
		pageCtx:       pageCtx,
		closer:        db,
		entries:       entries,
		mentions:      mentions,
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

	mux := route.New()
	mux.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, ErrNotFound) {
			slog.Error("not found", slog.String("url", r.URL.Path), slog.Any("err", err))
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		slog.Error("page error", slog.String("url", r.URL.Path), slog.Any("err", err))
		http.Error(w, "something unexpected happened", http.StatusInternalServerError)
	}

	mux.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) error {
		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.Before(before)
		if err != nil {
			return err
		}

		olderThan := ""
		if len(posts) == 25 {
			olderThan = posts[len(posts)-1].Properties["published"][0].(string)
		} else if len(posts) == 0 {
			olderThan = "NOMORE"
		}

		w.Header().Add("Link", `<`+indexURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)

		if _, err := page.List(b.pageCtx, page.ListData{
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
		}).WriteTo(w); err != nil {
			return err
		}

		return nil
	})

	mux.HandleFunc("/kind/:kind", func(w http.ResponseWriter, r *http.Request) error {
		vars := route.Vars(r)

		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.KindBefore(vars["kind"], before)
		if err != nil {
			return err
		}

		olderThan := ""
		if len(posts) == 25 {
			olderThan = posts[len(posts)-1].Properties["published"][0].(string)
		} else if len(posts) == 0 {
			olderThan = "NOMORE"
		}

		if _, err := page.List(b.pageCtx, page.ListData{
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
			Kind:         vars["kind"],
		}).WriteTo(w); err != nil {
			return err
		}

		return nil
	})

	mux.HandleFunc("/category/:category", func(w http.ResponseWriter, r *http.Request) error {
		vars := route.Vars(r)

		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.CategoryBefore(vars["category"], before)
		if err != nil {
			return err
		}

		olderThan := ""
		if len(posts) == 25 {
			olderThan = posts[len(posts)-1].Properties["published"][0].(string)
		} else if len(posts) == 0 {
			olderThan = "NOMORE"
		}

		if _, err := page.List(b.pageCtx, page.ListData{
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
			Kind:         "",
			Category:     vars["category"],
		}).WriteTo(w); err != nil {
			return fmt.Errorf("render: %w", err)
		}

		return nil
	})

	mux.HandleFunc("/entry/:id", func(w http.ResponseWriter, r *http.Request) error {
		vars := route.Vars(r)

		entry, err := b.EntryByUID(vars["id"])
		if err != nil {
			return fmt.Errorf("entry by uid: %w", err)
		}

		if deleted, ok := entry["hx-deleted"]; ok && len(deleted) > 0 {
			http.Error(w, "gone", http.StatusGone)
			return nil
		}

		mentions, err := b.MentionsForEntry(baseURL.ResolveReference(r.URL).String())
		if err != nil {
			return fmt.Errorf("mentions for entry: %w", err)
		}

		if _, err := page.Post(b.pageCtx, page.PostData{
			Entry: entry,
			Posts: GroupedPosts{
				Type: "entry",
				Meta: entry,
			},
			Mentions: mentions,
		}).WriteTo(w); err != nil {
			return fmt.Errorf("render: %w", err)
		}

		return nil
	})

	mux.HandleFunc("/likes/:ymd", func(w http.ResponseWriter, r *http.Request) error {
		ymd := route.Vars(r)["ymd"]

		likes, err := b.LikesOn(ymd)
		if err != nil {
			return err
		}

		if _, err := page.Day(b.pageCtx, page.DayData{
			Ymd:   ymd,
			Items: likes,
		}).WriteTo(w); err != nil {
			return fmt.Errorf("render: %w", err)
		}

		return nil
	})

	mux.HandleFunc("/mentions", func(w http.ResponseWriter, r *http.Request) error {
		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		mentions, err := b.MentionsBefore(before, 25)
		if err != nil {
			return err
		}

		olderThan := ""
		if len(mentions) == 25 {
			olderThan = mentions[len(mentions)-1].Properties["published"][0].(string)
		} else if len(mentions) == 0 {
			olderThan = "NOMORE"
		}

		if _, err := page.Mentions(b.pageCtx, page.MentionsData{
			Title:      "mentions",
			Items:      mentions,
			OlderThan:  olderThan,
			ShowLatest: showLatest,
		}).WriteTo(w); err != nil {
			return fmt.Errorf("render: %w", err)
		}

		return nil
	})

	mux.HandleFunc("/feed/rss", func(w http.ResponseWriter, r *http.Request) error {
		f, err := b.feed()
		if err != nil {
			return fmt.Errorf("get feed: %w", err)
		}

		rss, err := f.ToRss()
		if err != nil {
			return fmt.Errorf("to rss: %w", err)
		}

		w.Header().Add("Link", `<`+feedRssURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)
		w.Header().Set("Content-Type", "application/rss+xml")
		io.WriteString(w, rss)
		return nil
	})

	mux.HandleFunc("/feed/atom", func(w http.ResponseWriter, r *http.Request) error {
		f, err := b.feed()
		if err != nil {
			return fmt.Errorf("get feed: %w", err)
		}

		atom, err := f.ToAtom()
		if err != nil {
			return fmt.Errorf("to atom: %w", err)
		}

		w.Header().Add("Link", `<`+feedAtomURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)
		w.Header().Set("Content-Type", "application/atom+xml")
		io.WriteString(w, atom)
		return nil
	})

	mux.HandleFunc("/feed/jsonfeed", func(w http.ResponseWriter, r *http.Request) error {
		f, err := b.feed()
		if err != nil {
			return fmt.Errorf("get feed: %w", err)
		}

		json, err := f.ToJSON()
		if err != nil {
			return fmt.Errorf("to json: %w", err)
		}

		w.Header().Add("Link", `<`+feedJsonfeedURL+`>; rel="self"`)
		w.Header().Add("Link", `<`+b.config.HubURL+`>; rel="hub"`)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, json)
		return nil
	})

	// route.Handle("/:year/:month/:date/:slug")

	return mux
}

func (b *Blog) feed() (*feeds.Feed, error) {
	feed := &feeds.Feed{
		Title:   b.pageCtx.Name + " posts",
		Link:    &feeds.Link{Href: b.config.BaseURL.String()},
		Author:  &feeds.Author{Name: b.pageCtx.Name},
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
			Title:       page.DecideTitle(post.Properties),
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
