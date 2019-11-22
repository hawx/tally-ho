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
			fmt.Fprint(w, err)
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
			log.Println(err)
			return
		}

		mentions, err := b.DB.MentionsForEntry(r.URL.Path)
		if err != nil {
			log.Println(err)
			return
		}

		if err := b.Templates.ExecuteTemplate(w, "post.gotmpl", struct {
			Entry    map[string][]interface{}
			Mentions []numbersix.Group
		}{
			Entry:    entry,
			Mentions: mentions,
		}); err != nil {
			log.Println(err)
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
					log.Println("syndicating to twitter: ", err)
					continue
				}

				if err := b.Update(location, empty, map[string][]interface{}{
					"syndication": {syndicatedLocation},
				}, empty); err != nil {
					log.Println("updating with twitter location: ", err)
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
	Posts []numbersix.Group
	Meta  map[string][]interface{}
}

func groupLikes(posts []numbersix.Group) []GroupedPosts {
	var groupedPosts []GroupedPosts

	var today string
	var todaysLikes []numbersix.Group

	for _, post := range posts {
		if len(post.Properties["like-of"]) > 0 {
			likeDate := strings.Split(post.Properties["published"][0].(string), "T")[0]
			if likeDate == today {
				todaysLikes = append(todaysLikes, post)
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

				todaysLikes = []numbersix.Group{post}
				today = likeDate
			}
		} else {
			groupedPosts = append(groupedPosts, GroupedPosts{
				Type:  "entry",
				Posts: []numbersix.Group{post}},
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
