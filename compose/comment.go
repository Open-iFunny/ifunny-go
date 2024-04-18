package compose

import (
	"fmt"
	"net/url"
)

func Comments(id string, limit int, page Page[string]) Request {
	q := url.Values{"limit": []string{fmt.Sprint(limit)}}
	if page.Key != NONE {
		q.Set(string(page.Key), page.Value)
	}

	return get("/content/"+id+"/comments", q)
}
