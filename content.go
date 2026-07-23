package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

// ContentKind identifies the shape of a piece of Content: what was actually
// posted, and therefore which field on Content (Pic, VideoClip, Gif, ...)
// holds its payload. It's a defined string type rather than a bag of
// untyped constants so switches over it can be exhaustiveness-checked by
// linters such as exhaustive.
type ContentKind string

// Content kinds returned by the API. Cross-checked against
// github.com/MakeShiftArtist/ifunny-api-types (src/payloads/content.ts),
// which documents kinds beyond the three iFunny's own docs mention.
const (
	CONTENT_PIC         ContentKind = "pic"         // image meme; payload in Content.Pic
	CONTENT_COMICS      ContentKind = "comics"      // made with the in-app comic maker; payload in Content.Comics
	CONTENT_MEME        ContentKind = "mem"         // made with iFunny's legacy "Meme" creator; payload in Content.Meme
	CONTENT_VIDEO_CLIP  ContentKind = "video_clip"  // most common video kind; payload in Content.VideoClip
	CONTENT_VIDEO       ContentKind = "video"       // imported from a URL (e.g. YouTube); payload in Content.Video
	CONTENT_VINE        ContentKind = "vine"        // legacy Vine import; payload in Content.Vine
	CONTENT_COUB        ContentKind = "coub"        // imported from coub.com; payload in Content.Coub
	CONTENT_GIF         ContentKind = "gif"         // plain gif; payload in Content.Gif
	CONTENT_GIF_CAPTION ContentKind = "gif_caption" // captioned gif; shares Content.Gif with CONTENT_GIF
	CONTENT_CAPTION     ContentKind = "caption"     // captioned image meme; payload in Content.Caption
	CONTENT_APP         ContentKind = "app"         // interactive/iframe card, deprecated by iFunny; payload in Content.App
	CONTENT_OLD         ContentKind = "old"         // ? undocumented, deprecated; no known payload
	CONTENT_DEM         ContentKind = "dem"         // ? undocumented, deprecated; no known payload
	CONTENT_SPECIAL     ContentKind = "special"     // ? undocumented; no known payload
)

// Content represents a post (image, video, gif, ...) on iFunny. It includes
// metadata like creation date, stats (smiles, comments), and the creator's
// information. The kind-specific data describing what was actually posted
// lives in one of the Pic/VideoClip/Gif/... fields, selected by Type; use
// Payload to fetch it generically.
type Content struct {
	Type        ContentKind `json:"type"`
	ID          string      `json:"id"`
	Link        string      `json:"link"`
	DateCreated int64       `json:"date_created"`
	PublushAt   int64       `json:"publish_at"`
	Tags        []string    `json:"tags"`
	State       string      `json:"state"`
	ShotStatus  string      `json:"shot_status"`

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

	// Populated when Type == CONTENT_PIC.
	Pic *PicPayload `json:"pic,omitempty"`
	// Populated when Type == CONTENT_COMICS. Same shape as Pic; comics are
	// just images produced by a different in-app creator.
	Comics *PicPayload `json:"comics,omitempty"`
	// Populated when Type == CONTENT_MEME. Same shape as Pic; memes come
	// from iFunny's legacy "Meme" creator.
	Meme *PicPayload `json:"mem,omitempty"`
	// Populated when Type == CONTENT_VIDEO_CLIP.
	VideoClip *VideoClipPayload `json:"video_clip,omitempty"`
	// Populated when Type == CONTENT_VIDEO.
	Video *VideoPayload `json:"video,omitempty"`
	// Populated when Type == CONTENT_VINE.
	Vine *VinePayload `json:"vine,omitempty"`
	// Populated when Type == CONTENT_COUB.
	Coub *CoubPayload `json:"coub,omitempty"`
	// Populated when Type is CONTENT_GIF or CONTENT_GIF_CAPTION. The API
	// uses the same "gif" key for both; CaptionText is only meaningful
	// (non-empty) for CONTENT_GIF_CAPTION.
	Gif *GifPayload `json:"gif,omitempty"`
	// Populated when Type == CONTENT_CAPTION.
	Caption *CaptionPayload `json:"caption,omitempty"`
	// Populated when Type == CONTENT_APP.
	App *AppPayload `json:"app,omitempty"`
}

// Payload returns the kind-specific data for c as a Payload, chosen by
// c.Type. It returns nil for kinds with no documented payload (CONTENT_OLD,
// CONTENT_DEM, CONTENT_SPECIAL) or if the expected field wasn't populated.
//
// Callers that don't care which exact kind they're holding can type-assert
// the result against the narrower capability interfaces below (Media,
// Preview, Sized, Captioned) instead of switching over every ContentKind:
//
//	if media, ok := content.Payload().(ifunny.Media); ok {
//		fmt.Println(media.URL())
//	}
func (c *Content) Payload() Payload {
	switch c.Type {
	case CONTENT_PIC:
		return payloadOrNil(c.Pic)
	case CONTENT_COMICS:
		return payloadOrNil(c.Comics)
	case CONTENT_MEME:
		return payloadOrNil(c.Meme)
	case CONTENT_VIDEO_CLIP:
		return payloadOrNil(c.VideoClip)
	case CONTENT_VIDEO:
		return payloadOrNil(c.Video)
	case CONTENT_VINE:
		return payloadOrNil(c.Vine)
	case CONTENT_COUB:
		return payloadOrNil(c.Coub)
	case CONTENT_GIF, CONTENT_GIF_CAPTION:
		return payloadOrNil(c.Gif)
	case CONTENT_CAPTION:
		return payloadOrNil(c.Caption)
	case CONTENT_APP:
		return payloadOrNil(c.App)
	default:
		return nil
	}
}

// payloadOrNil returns p as a Payload, or a true nil interface (rather than
// a non-nil interface wrapping a nil pointer) when p itself is nil. The
// comparison happens before p is boxed into the Payload interface, so it
// sidesteps the usual "non-nil interface holding a nil pointer" trap.
func payloadOrNil[T interface {
	Payload
	comparable
}](p T) Payload {
	var zero T
	if p == zero {
		return nil
	}
	return p
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
