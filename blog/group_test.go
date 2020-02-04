package blog

import (
	"testing"

	"hawx.me/code/assert"
	"hawx.me/code/numbersix"
)

func TestGroupLikesJustPosts(t *testing.T) {
	posts := []numbersix.Group{
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"published": {"2019-01-01T14:00:00Z"},
			},
		},
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"published": {"2019-02-01T14:00:00Z"},
			},
		},
	}

	grouped := groupLikes(posts)

	if assert.Len(t, grouped, 2) {
		assert.Equal(t, GroupedPosts{
			Type: "entry",
			Meta: posts[1].Properties,
		}, grouped[0])

		assert.Equal(t, GroupedPosts{
			Type: "entry",
			Meta: posts[0].Properties,
		}, grouped[1])
	}
}

func TestGroupLikesJustLikes(t *testing.T) {
	posts := []numbersix.Group{
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"hx-kind":   {"like"},
				"published": {"2019-01-01T14:00:00Z"},
			},
		},
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"hx-kind":   {"like"},
				"published": {"2019-02-01T14:00:00Z"},
			},
		},
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"hx-kind":   {"like"},
				"published": {"2019-02-01T17:00:00Z"},
			},
		},
	}

	grouped := groupLikes(posts)

	if assert.Len(t, grouped, 2) {
		assert.Equal(t, GroupedPosts{
			Type: "like",
			Posts: []map[string][]interface{}{
				posts[1].Properties,
				posts[2].Properties,
			},
			Meta: map[string][]interface{}{
				"url":       {"/likes/2019-02-01"},
				"published": {"2019-02-01T12:00:00Z"},
			},
		}, grouped[0])

		assert.Equal(t, GroupedPosts{
			Type: "like",
			Posts: []map[string][]interface{}{
				posts[0].Properties,
			},
			Meta: map[string][]interface{}{
				"url":       {"/likes/2019-01-01"},
				"published": {"2019-01-01T12:00:00Z"},
			},
		}, grouped[1])
	}
}

func TestGroupLikesMixed(t *testing.T) {
	posts := []numbersix.Group{
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"hx-kind":   {"like"},
				"published": {"2019-01-01T14:00:00Z"},
			},
		},
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"published": {"2019-01-01T15:00:00Z"},
			},
		},
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"hx-kind":   {"like"},
				"published": {"2019-02-01T14:00:00Z"},
			},
		},

		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"published": {"2019-02-01T15:00:00Z"},
			},
		},
		{
			Subject: "1",
			Properties: map[string][]interface{}{
				"hx-kind":   {"like"},
				"published": {"2019-02-01T17:00:00Z"},
			},
		},
	}

	grouped := groupLikes(posts)

	if assert.Len(t, grouped, 4) {
		assert.Equal(t, GroupedPosts{
			Type: "entry",
			Meta: posts[3].Properties,
		}, grouped[0])

		assert.Equal(t, GroupedPosts{
			Type: "like",
			Posts: []map[string][]interface{}{
				posts[2].Properties,
				posts[4].Properties,
			},
			Meta: map[string][]interface{}{
				"url":       {"/likes/2019-02-01"},
				"published": {"2019-02-01T12:00:00Z"},
			},
		}, grouped[1])

		assert.Equal(t, GroupedPosts{
			Type: "entry",
			Meta: posts[1].Properties,
		}, grouped[2])

		assert.Equal(t, GroupedPosts{
			Type: "like",
			Posts: []map[string][]interface{}{
				posts[0].Properties,
			},
			Meta: map[string][]interface{}{
				"url":       {"/likes/2019-01-01"},
				"published": {"2019-01-01T12:00:00Z"},
			},
		}, grouped[3])
	}
}
