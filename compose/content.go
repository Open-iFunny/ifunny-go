package compose

import (
	"fmt"
	"net/url"
)

func ContentByID(id string) Request {
	return get("/content/"+id, nil)
}

func Feed(feed string, limit int, page Page[string]) Request {
	q := url.Values{"limit": []string{fmt.Sprint(limit)}}
	if feed == "collective" {
		return Request{"POST", "/feeds/collective", nil, q}
	}

	return get("/feeds/"+feed, q)
}

func Timeline(id string, limit int, page Page[string]) Request {
	return Feed("/timelines/users/"+id, limit, page)
}
