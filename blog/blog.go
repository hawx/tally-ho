package blog

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"hawx.me/code/numbersix"
	"hawx.me/code/route"
)

type Syndicator interface {
	Create(map[string][]interface{}) (string, error)
}

type Blog struct {
	Me          string
	Name        string
	Title       string
	Description string
	DB          *DB
	Templates   *template.Template
	Twitter     Syndicator
}

func (b *Blog) Handler() http.Handler {
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

		mentions, err := b.DB.MentionsForEntry(r.URL.Path)
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
				Type:  "entry",
				Posts: []map[string][]interface{}{entry},
			},
			Mentions: mentions,
		}); err != nil {
			log.Printf("ERR get-entry-render url=%s; %v\n", r.URL.Path, err)
		}
	})

	// route.Handle("/:year/:month/:date/:slug")
	// route.Handle("/like/...")
	// route.Handle("/reply/...")
	// route.Handle("/tag/...")

	return route.Default
}

func (b *Blog) Entry(url string) (data map[string][]interface{}, err error) {
	return b.DB.Entry(url)
}

func (b *Blog) Create(data map[string][]interface{}) (location string, err error) {
	location, err = b.DB.Create(data)
	if err != nil {
		return
	}

	if syndicateTos, ok := data["mp-syndicate-to"]; ok && len(syndicateTos) > 0 {
		for _, syndicateTo := range syndicateTos {
			if syndicateTo == "https://twitter.com/" {
				syndicatedLocation, err := b.Twitter.Create(data)
				if err != nil {
					log.Printf("ERR syndication to=twitter uid=%s; %v\n", data["uid"][0], err)
					continue
				}

				if err := b.Update(location, empty, map[string][]interface{}{
					"syndication": {syndicatedLocation},
				}, empty); err != nil {
					log.Printf("ERR confirming-syndication to=twitter uid=%s; %v\n", data["uid"][0], err)
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
				Type:  "entry",
				Posts: []map[string][]interface{}{post.Properties}},
			)
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

	return groupedPosts
}
