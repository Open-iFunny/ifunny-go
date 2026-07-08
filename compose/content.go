package compose

import (
	"fmt"
	"net/url"
)

// ContentByID composes a request for content by ID.
func ContentByID(id string) Request {
	return get("/content/"+id, nil)
}

func feedParams(limit int, page Page[string]) url.Values {
	q := url.Values{"limit": []string{fmt.Sprint(limit)}}
	if page.Key != NONE {
		q.Set(string(page.Key), page.Value)
	}
	return q
}

// Feed composes a request for a named content feed (e.g. "featured", "trending") with pagination.
func Feed(feed string, limit int, page Page[string]) Request {
	if feed == "collective" {
		return Request{"POST", "/feeds/collective", nil, feedParams(limit, page)}
	}

	return get("/feeds/"+feed, feedParams(limit, page))
}

// Timeline composes a request for a user's content timeline by ID with pagination.
func Timeline(id string, limit int, page Page[string]) Request {
	return get("/timelines/users/"+id, feedParams(limit, page))
}

// TimelineByNick composes a request for a user's content timeline by nick with pagination.
func TimelineByNick(nick string, limit int, page Page[string]) Request {
	return get("/timelines/users/by_nick/"+nick, feedParams(limit, page))
}

// Smiles composes a request for users who smiled on a content item with pagination.
func Smiles(id string, limit int, page Page[string]) Request {
	return get("/content/"+id+"/smiles", feedParams(limit, page))
}

// Republished composes a request for users who republished a content item with pagination.
func Republished(id string, limit int, page Page[string]) Request {
	return get("/content/"+id+"/republished", feedParams(limit, page))
}

/*
content_top_today
content_top_this_week
content_top_year_{2024..2018}
content_top_month_{january..december}
content_top_overall
content_top_by_share
content_shuffle

channel-wtf
channel-animals
channel-games
channel-comic
channel-video
channel-sports
channel-ifunny-originals
channel-wholesome-wednesday

category-animals-nature
category-anime-manga
category-art-creative
category-cars
category-celebrities
category-gaming
category-girls
category-internet
category-memes
category-movies
category-other
category-politics
category-science-tech

users_top_by_featured
users_top_by_subscribers
users_top_by_content_views
users_top_overall

chats_popular_last_week
chats_new_chats
chats_top_by_members
*/
// Explore composes a request for content from an explore compilation by ID with pagination.
func Explore(id string, limit int, page Page[string]) Request {
	return Request{"POST", "/explore/compilation/" + id, nil, feedParams(limit, page)}
}
