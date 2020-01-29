package blog

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
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
	BaseURL     *url.URL
	MediaURL    *url.URL
}

type Blog struct {
	Config      Config
	MediaDir    string
	DB          *DB
	Templates   *template.Template
	Syndicators map[string]syndicate.Syndicator
}

func (b *Blog) BaseURL() string {
	return b.Config.BaseURL.String()
}

func (b *Blog) Handler() http.Handler {
	baseURL := b.Config.BaseURL

	route.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		showLatest := true

		before, err := time.Parse(time.RFC3339, r.FormValue("before"))
		if err != nil {
			showLatest = false
			before = time.Now().UTC()
		}

		posts, err := b.DB.Before(before)
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

		if err := b.Templates.ExecuteTemplate(w, "list.gotmpl", struct {
			GroupedPosts []GroupedPosts
			OlderThan    string
			ShowLatest   bool
		}{
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

		posts, err := b.DB.KindBefore(vars["kind"], before)
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

		if err := b.Templates.ExecuteTemplate(w, "list.gotmpl", struct {
			GroupedPosts []GroupedPosts
			OlderThan    string
			ShowLatest   bool
		}{
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
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

		posts, err := b.DB.CategoryBefore(vars["category"], before)
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

		if err := b.Templates.ExecuteTemplate(w, "list.gotmpl", struct {
			GroupedPosts []GroupedPosts
			OlderThan    string
			ShowLatest   bool
		}{
			GroupedPosts: groupLikes(posts),
			OlderThan:    olderThan,
			ShowLatest:   showLatest,
		}); err != nil {
			fmt.Fprint(w, err)
		}
	})

	route.HandleFunc("/entry/:id", func(w http.ResponseWriter, r *http.Request) {
		entry, err := b.DB.Entry(r.URL.Path)
		if err != nil {
			log.Printf("ERR get-entry url=%s; %v\n", r.URL.Path, err)
			return
		}

		if deleted, ok := entry["hx-deleted"]; ok && len(deleted) > 0 {
			http.Error(w, "gone", http.StatusGone)
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
	if strings.HasPrefix(url, b.BaseURL()) {
		url = url[len(b.BaseURL()):]
		if url[0] != '/' {
			url = "/" + url
		}
	}

	return b.DB.Entry(url)
}

func (b *Blog) Update(
	url string,
	replace, add, delete map[string][]interface{},
	deleteAll []string,
) error {
	if !strings.HasPrefix(url, b.BaseURL()) {
		return errors.New("expected url to be for this blog")
	}
	url = url[len(b.BaseURL()):]
	if url[0] != '/' {
		url = "/" + url
	}

	return b.DB.Update(url, replace, add, delete, deleteAll)
}

func (b *Blog) Delete(url string) error {
	if !strings.HasPrefix(url, b.BaseURL()) {
		return errors.New("expected url to be for this blog")
	}
	url = url[len(b.BaseURL()):]
	if url[0] != '/' {
		url = "/" + url
	}

	return b.DB.Delete(url)
}

func (b *Blog) Undelete(url string) error {
	if !strings.HasPrefix(url, b.BaseURL()) {
		return errors.New("expected url to be for this blog")
	}
	url = url[len(b.BaseURL()):]
	if url[0] != '/' {
		url = "/" + url
	}

	return b.DB.Undelete(url)
}

func (b *Blog) Mention(source string, data map[string][]interface{}) error {
	return b.DB.Mention(source, data)
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
