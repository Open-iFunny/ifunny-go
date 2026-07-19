package compose

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Placement selects where a page token rides on the request: as a URL query
// param (the historical behavior, and the only option for GET feeds) or in the
// urlencoded POST body. Body placement exists for the collective feed, whose
// cursor grows large enough to blow a server-side query-length limit.
type Placement int

const (
	InQuery Placement = iota
	InBody
)

const defaultLimit = 30

// Feed describes a paginated endpoint in a way that reconciles collective,
// featured/named feeds, timelines, comments, subscribers and chat search. Build
// one page's request with [Feed.Request]; drive iteration by feeding the
// server's returned cursor back through Pager.
type Feed struct {
	Path   string     // request path, e.g. "/feeds/collective"
	Method string     // HTTP method; "" means GET
	Limit  int        // page size; 0 means defaultLimit (30)
	Params url.Values // extra static query params (e.g. chat search "q")

	// Pager transforms the raw cursor the server returned into the token sent
	// on the next request. nil means echo the server's cursor verbatim (the
	// historical behavior). For collective this is a [TailPager].
	Pager func(string) string

	// PagerIn selects query vs body placement of the page token. Body placement
	// only takes effect on POST feeds; it is ignored (treated as InQuery) for GET.
	PagerIn Placement
}

// Request builds the request for a single page. limit and any static Params go
// on the query string as before; the page token goes on the query string, or —
// for a POST feed with PagerIn==InBody — into a urlencoded body with the
// matching Content-Type.
func (f Feed) Request(page Page) Request {
	method := f.Method
	if method == "" {
		method = "GET"
	}

	limit := f.Limit
	if limit == 0 {
		limit = defaultLimit
	}

	query := url.Values{}
	for k, v := range f.Params {
		query[k] = v
	}
	query.Set("limit", strconv.Itoa(limit))

	req := Request{Method: method, Path: f.Path, Query: query}
	if page.Key == NONE {
		return req
	}

	if f.PagerIn == InBody && method == "POST" {
		body := url.Values{string(page.Key): {page.Value.String()}}
		req.Body = strings.NewReader(body.Encode())
		req.Header = http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}
	} else {
		query.Set(string(page.Key), page.Value.String())
	}

	return req
}

// NamedFeed describes a named content feed (e.g. "featured", "collective").
// Use [Collective] for collective when you want body placement and tail-paging.
func NamedFeed(name string) Feed {
	return Feed{Path: "/feeds/" + name}
}

// Collective describes the collective feed with the mitigations proven to dodge
// its pagination size cliff: the cursor is posted in the urlencoded body, and
// each outgoing token is reduced to the last `tail` seen IDs. tail <= 0 disables
// tail-paging (echo the server cursor verbatim), matching NamedFeed("collective")
// but still in the POST body.
func Collective(tail int) Feed {
	return Feed{
		Path:    "/feeds/collective",
		Method:  "POST",
		Pager:   TailPager(tail),
		PagerIn: InBody,
	}
}

// Timeline describes a user's content timeline by ID.
func Timeline(id string) Feed { return Feed{Path: "/timelines/users/" + id} }

// TimelineByNick describes a user's content timeline by nick.
func TimelineByNick(nick string) Feed { return Feed{Path: "/timelines/users/by_nick/" + nick} }

// Smiles describes the users who smiled on a content item.
func Smiles(id string) Feed { return Feed{Path: "/content/" + id + "/smiles"} }

// Republished describes the users who republished a content item.
func Republished(id string) Feed { return Feed{Path: "/content/" + id + "/republished"} }

// Explore describes content from an explore compilation by ID.
func Explore(id string) Feed { return Feed{Path: "/explore/compilation/" + id, Method: "POST"} }

// Comments describes the comments on a content item.
func Comments(id string) Feed { return Feed{Path: "/content/" + id + "/comments"} }

// Replies describes the replies to a comment on a content item.
func Replies(cid, id string) Feed {
	return Feed{Path: "/content/" + cid + "/comments/" + id + "/replies"}
}

// Subscribers describes a user's subscribers.
func Subscribers(id string) Feed { return Feed{Path: "/users/" + id + "/subscribers"} }

// Subscriptions describes a user's subscriptions.
func Subscriptions(id string) Feed { return Feed{Path: "/users/" + id + "/subscriptions"} }

// Chats describes an open-channel search by query string.
func Chats(query string) Feed {
	return Feed{Path: "/chats/open_channels", Params: url.Values{"q": {query}}}
}
