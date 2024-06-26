package ifunny

import (
	"github.com/open-ifunny/ifunny-go/compose"
)

// TODO types of content, missing a bunch of content types
// an enum type when we have all of them ( will we ever? )
const (
	CONTENT_PIC        = "pic"
	CONTENT_VIDEO_CLIP = "video_clip"
	CONTENT_COMICS     = "comics" // created from the in-app comic maker
)

type Content struct {
	Type        string   `json:"type"`
	ID          string   `json:"id"`
	Link        string   `json:"link"`
	DateCreated int64    `json:"date_created"`
	PublushAt   int64    `json:"publish_at"`
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

type Cursor struct {
	Cursors struct {
		Next string `json:"next,omitempty"`
		Prev string `json:"prev,omitempty"`
	} `json:"cursors"`
	HasNext bool `json:"hasNext"`
	HasPrev bool `json:"hasPrev"`
}

func (client *Client) GetContent(id string) (*Content, error) {
	content := new(struct {
		Data Content `json:"data"`
	})
	err := client.RequestJSON(compose.ContentByID(id), content)
	return &content.Data, err
}

func (client *Client) GetFeedPage(request compose.Request) (*Page[Content], error) {
	content := new(struct {
		Data struct {
			Content Page[Content] `json:"content"`
		} `json:"data"`
	})

	err := client.RequestJSON(request, content)
	return &content.Data.Content, err
}

// func (client *Client) IterFeed(feed string) Iterator[*Content] {

// }

func (client *Client) IterFeed(feed string) <-chan Result[*Content] {
	return iterFrom(client, func(limit int, page compose.Page[string]) compose.Request { return compose.Feed(feed, limit, page) }, client.GetFeedPage)
}

func (client *Client) IterTimeline(id string) <-chan Result[*Content] {
	return iterFrom(client, func(limit int, page compose.Page[string]) compose.Request { return compose.Timeline(id, limit, page) }, client.GetFeedPage)
}
