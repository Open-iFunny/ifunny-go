package compose

// ContentByID composes a request for content by ID.
func ContentByID(id string) Request {
	return get("/content/"+id, nil)
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
