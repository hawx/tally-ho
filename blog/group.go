package blog

import (
	"sort"
	"strings"

	"hawx.me/code/numbersix"
	"hawx.me/code/tally-ho/internal/page"
)

type GroupedPosts = page.GroupedPosts

func groupLikes(ctx page.Context, posts []numbersix.Group) []GroupedPosts {
	var groupedPosts []GroupedPosts

	var today string
	var todaysLikes []map[string][]any

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Properties["published"][0].(string) < posts[j].Properties["published"][0].(string)
	})

	for _, post := range posts {
		if kind, ok := post.Properties["hx-kind"]; ok && len(kind) > 0 && kind[0] == "like" {
			likeDate := strings.Split(post.Properties["published"][0].(string), "T")[0]
			if likeDate == today {
				todaysLikes = append(todaysLikes, post.Properties)
			} else {
				if len(todaysLikes) > 0 {
					groupedPosts = append(groupedPosts, GroupedPosts{
						Type:  "like",
						Posts: todaysLikes,
						Meta: map[string][]any{
							"url":       {ctx.Path("likes/" + today)},
							"published": {today + "T12:00:00Z"},
						},
					})
				}

				todaysLikes = []map[string][]any{post.Properties}
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
			Meta: map[string][]any{
				"url":       {ctx.Path("likes/" + today)},
				"published": {today + "T12:00:00Z"},
			},
		})
	}

	sort.Slice(groupedPosts, func(i, j int) bool {
		return groupedPosts[i].Meta["published"][0].(string) > groupedPosts[j].Meta["published"][0].(string)
	})

	return groupedPosts
}
