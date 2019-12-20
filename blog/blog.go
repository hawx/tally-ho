package blog

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"hawx.me/code/numbersix"
	"hawx.me/code/route"
	"hawx.me/code/tally-ho/syndicate"
)

type Config struct {
	Me          string
	Name        string
	Title       string
	Description string
	BaseURL     string
}

type Blog struct {
	Config      Config
	DB          *DB
	Templates   *template.Template
	Syndicators []syndicate.Syndicator
}

func (b *Blog) BaseURL() string {
	return b.Config.BaseURL
}

func (b *Blog) Handler() http.Handler {
	baseURL, _ := url.Parse(b.Config.BaseURL)

	route.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		posts, err := b.DB.Before(time.Now().UTC())
		if err != nil {
			log.Println("ERR get-all;", err)
			return
		}

		groupedPosts := groupLikes(posts)

		if err := b.Templates.ExecuteTemplate(w, "list.gotmpl", groupedPosts); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/entry/:id", func(w http.ResponseWriter, r *http.Request) {
		entry, err := b.DB.Entry(r.URL.Path)
		if err != nil {
			log.Printf("ERR get-entry url=%s; %v\n", r.URL.Path, err)
			return
		}

		mentions, err := b.DB.MentionsForEntry(baseURL.ResolveReference(r.URL).String())
		if err != nil {
			log.Printf("ERR get-entry-mentions url=%s; %v\n", r.URL.Path, err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "post.gotmpl", struct {
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

		likes, err := b.DB.LikesOn(ymd)
		if err != nil {
			log.Printf("ERR likes-on ymd=%s; %v\n", ymd, err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "day.gotmpl", likes); err != nil {
			log.Printf("ERR likes-on-render ymd=%s; %v\n", ymd, err)
		}
	})

	// route.Handle("/:year/:month/:date/:slug")
	// route.Handle("/like/...")
	// route.Handle("/reply/...")
	// route.Handle("/tag/...")

	return route.Default
}

func (b *Blog) Entry(url string) (data map[string][]interface{}, err error) {
	if strings.HasPrefix(url, b.Config.BaseURL) {
		url = url[len(b.Config.BaseURL):]
		if url[0] != '/' {
			url = "/" + url
		}
	}

	return b.DB.Entry(url)
}

func (b *Blog) Create(data map[string][]interface{}) (location string, err error) {
	if len(data["name"]) == 0 {
		name, err := getName(data)
		if err != nil {
			log.Printf("WARN get-name; %v\n", err)
		} else if name != "" {
			data["name"] = []interface{}{name}
			log.Printf("WARN get-name; setting to '%s'\n", name)
		}
	}

	location, err = b.DB.Create(data)
	if err != nil {
		return
	}

	if syndicateTos, ok := data["mp-syndicate-to"]; ok && len(syndicateTos) > 0 {
		for _, syndicateTo := range syndicateTos {
			for _, syndicator := range b.Syndicators {
				if syndicateTo == syndicator.Config().UID {
					syndicatedLocation, err := syndicator.Create(data)
					if err != nil {
						log.Printf("ERR syndication to=%s uid=%s; %v\n", syndicator.Config().Name, data["uid"][0], err)
						continue
					}

					if err := b.Update(location, empty, map[string][]interface{}{
						"syndication": {syndicatedLocation},
					}, empty); err != nil {
						log.Printf("ERR confirming-syndication to=%s uid=%s; %v\n", syndicator.Config().Name, data["uid"][0], err)
					}
				}
			}
		}
	}

	return
}

func (b *Blog) Update(url string, replace, add, delete map[string][]interface{}) error {
	return b.DB.Update(url, replace, add, delete)
}

func (b *Blog) Mention(source string, data map[string][]interface{}) error {
	return b.DB.Mention(source, data)
}

type GroupedPosts struct {
	Type  string
	Posts []map[string][]interface{}
	Meta  map[string][]interface{}
}

func groupLikes(posts []numbersix.Group) []GroupedPosts {
	var groupedPosts []GroupedPosts

	var today string
	var todaysLikes []map[string][]interface{}

	for _, post := range posts {
		if len(post.Properties["like-of"]) > 0 {
			likeDate := strings.Split(post.Properties["published"][0].(string), "T")[0]
			if likeDate == today {
				todaysLikes = append(todaysLikes, post.Properties)
			} else {
				if len(todaysLikes) > 0 {
					groupedPosts = append(groupedPosts, GroupedPosts{
						Type:  "like",
						Posts: todaysLikes,
						Meta: map[string][]interface{}{
							"url":       {"/likes/" + today},
							"published": {today + "T12:00:00Z"},
						},
					})
				}

				todaysLikes = []map[string][]interface{}{post.Properties}
				today = likeDate
			}
		} else {
			groupedPosts = append(groupedPosts, GroupedPosts{
				Type: "entry",
				Meta: post.Properties,
			})
		}
	}

	if len(todaysLikes) > 0 {
		groupedPosts = append(groupedPosts, GroupedPosts{
			Type:  "like",
			Posts: todaysLikes,
			Meta: map[string][]interface{}{
				"url":       {"/likes/" + today},
				"published": {today + "T12:00:00Z"},
			},
		})
	}

	sort.Slice(groupedPosts, func(i, j int) bool {
		return groupedPosts[i].Meta["published"][0].(string) > groupedPosts[j].Meta["published"][0].(string)
	})

	return groupedPosts
}

func getName(data map[string][]interface{}) (string, error) {
	if len(data["like-of"]) > 0 {
		likeOf := data["like-of"][0].(string)

		resp, err := http.Get(likeOf)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		root, err := html.Parse(resp.Body)
		if err != nil {
			return "", err
		}

		hentries := searchAll(root, func(node *html.Node) bool {
			return node.Type == html.ElementNode && hasAttr(node, "class", "h-entry")
		})

		for _, hentry := range hentries {
			names := searchAll(hentry, func(node *html.Node) bool {
				return node.Type == html.ElementNode && hasAttr(node, "class", "p-name")
			})

			if len(names) > 0 {
				return textOf(names[0]), nil
			}
		}

		titles := searchAll(root, func(node *html.Node) bool {
			return node.Type == html.ElementNode && node.DataAtom == atom.Title
		})

		if len(titles) > 0 {
			return textOf(titles[0]), nil
		}
	}

	return "", errors.New("no name to find")
}
