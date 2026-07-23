package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

// TODO types of content, missing a bunch of content types
// an enum type when we have all of them ( will we ever? )
// Content types returned by the API.
const (
	CONTENT_PIC        = "pic"        // Picture/image content
	CONTENT_VIDEO_CLIP = "video_clip" // Video content
	CONTENT_COMICS     = "comics"     // created from the in-app comic maker
)

// Content represents a post (image, video, or comic) on iFunny. It includes metadata
// like creation date, stats (smiles, comments), and the creator's information.
type Content struct {
	Type        string   `json:"type"`
	ID          string   `json:"id"`
	Link        string   `json:"link"`
	DateCreated int64    `json:"date_created"`
	PublishAt   int64    `json:"publish_at"`
	IssueAt     int64    `json:"issue_at"`
	Tags        []string `json:"tags"`
	State       string   `json:"state"`
	ShotStatus  string   `json:"shot_status"`

	FastStart  bool `json:"fast_start"`
	IsFeatured bool `json:"is_featured"`
	IsPinned   bool `json:"is_pinned"`
	IsAbused   bool `json:"is_abused"`
	IsUnsafe   bool `json:"is_unsafe"`

	IsRepublished bool `json:"is_republished"`
	IsSmiled      bool `json:"is_smiled"`
	IsUnsmiled    bool `json:"is_unsmiled"`

	Size struct {
		Height int `json:"h"`
		Width  int `json:"w"`
	} `json:"size"`

	Num struct {
		Comments    int `json:"comments"`
		Republished int `json:"republished"`
		Smiles      int `json:"smiles"`
		Unsmiles    int `json:"unsmiles"`
		Views       int `json:"views"`
	} `json:"num"`

	Creator struct {
		ID   string `json:"id"`
		Nick string `json:"nick"`
	} `json:"creator"`

	// type = pic|comics
	Pic struct {
		WebpURL string `json:"webp_url"`
	} `json:"pic,omitempty"`

	// type = video_clip
	VideoClip struct {
		ScreenURL  string `json:"screen_url"`
		SourceType string `json:"source_type"`
		Bytes      int    `json:"bytes"`
		Duration   int    `json:"duration"`
	} `json:"video_clip,omitempty"`
}

// Cursor holds pagination metadata for paginated API responses. It includes opaque
// cursor strings for the next and previous pages, and flags indicating whether those
// pages exist.
type Cursor struct {
	Cursors struct {
		Next string `json:"next,omitempty"`
		Prev string `json:"prev,omitempty"`
	} `json:"cursors"`
	HasNext bool `json:"hasNext"`
	HasPrev bool `json:"hasPrev"`
}

// GetContent fetches a single piece of content by ID.
func (client *Client) GetContent(ctx context.Context, id string) (*Content, error) {
	content := new(struct {
		Data Content `json:"data"`
	})
	err := client.RequestJSON(ctx, compose.ContentByID(id), content)
	return &content.Data, err
}

// GetFeedPage fetches a single page of content given a composed request. It is used
// internally by feed iteration methods and is exported for advanced use cases.
func (client *Client) GetFeedPage(ctx context.Context, request compose.Request) (*Page[Content], error) {
	content := new(struct {
		Data struct {
			Content Page[Content] `json:"content"`
		} `json:"data"`
	})

	err := client.RequestJSON(ctx, request, content)
	return &content.Data.Content, err
}

// IterContent returns a channel that yields content from an arbitrary feed
// descriptor. It is the generic entry point behind IterTimeline/IterCollective/etc.,
// and the way to iterate a named feed (see [compose.NamedFeed]) or collective with
// custom knobs (see [compose.Collective]). The iterator automatically fetches new
// pages as needed. Cancel ctx to stop.
func (client *Client) IterContent(ctx context.Context, feed compose.Feed) <-chan Result[*Content] {
	return iterFrom(ctx, client, feed, client.GetFeedPage)
}

// IterCollective returns a channel that yields content from the collective feed,
// posting the cursor in the request body and reducing each outgoing token to the
// last `tail` seen IDs to dodge the collective pagination size cliff. A tail of 0
// disables tail-paging (verbatim cursor) while keeping body placement.
func (client *Client) IterCollective(ctx context.Context, tail int) <-chan Result[*Content] {
	return client.IterContent(ctx, compose.Collective(tail))
}

// IterTimeline returns a channel that yields content posted by a user (identified by ID).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterTimeline(ctx context.Context, id string) <-chan Result[*Content] {
	return client.IterContent(ctx, compose.Timeline(id))
}

// IterTimelineByNick returns a channel that yields content posted by a user (identified by nick/username).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterTimelineByNick(ctx context.Context, nick string) <-chan Result[*Content] {
	return client.IterContent(ctx, compose.TimelineByNick(nick))
}

// IterSmiles returns a channel that yields users who smiled at the content (identified by ID).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterSmiles(ctx context.Context, id string) <-chan Result[*User] {
	return iterFrom(ctx, client, compose.Smiles(id), client.GetUsersPage)
}

// IterRepublishers returns a channel that yields users who republished the content (identified by ID).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterRepublishers(ctx context.Context, id string) <-chan Result[*User] {
	return iterFrom(ctx, client, compose.Republished(id), client.GetUsersPage)
}
