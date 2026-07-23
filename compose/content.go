package compose

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ContentByID composes a request for content by ID.
func ContentByID(id string) Request {
	return get("/content/"+id, nil)
}

// patchContent composes a urlencoded PATCH against a content item, the wire
// form the app uses to edit pending delayed ("timer") posts.
func patchContent(id string, form url.Values) Request {
	return Request{
		Method: "PATCH",
		Path:   "/content/" + id,
		Body:   strings.NewReader(form.Encode()),
		Header: http.Header{"Content-Type": {"application/x-www-form-urlencoded"}},
	}
}

// ContentSchedule composes a request that moves the publish timer of a pending
// delayed post. publishAt is a unix timestamp in seconds.
func ContentSchedule(id string, publishAt int64) Request {
	return patchContent(id, url.Values{"publish_at": {strconv.FormatInt(publishAt, 10)}})
}

// ContentVisibility composes a request that changes the visibility of a
// pending delayed post (see VISIBILITY_* in the root package).
func ContentVisibility(id, visibility string) Request {
	return patchContent(id, url.Values{"visibility": {visibility}})
}

/*
Known explore-compilation and feed IDs (for use with [Explore] / [NamedFeed]):

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
