package blog

import (
	"sort"
	"strings"

	"hawx.me/code/numbersix"
)

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
