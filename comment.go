package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

// Comment represents a user's comment on content. It includes the comment text,
// author information, engagement metrics (smiles/unsmiles), and threading metadata
// (parent, depth, etc.).
type Comment struct {
	ID           string `json:"id"`
	CID          string `json:"cid"`
	State        string `json:"state"`
	Date         int    `json:"date"`
	Text         string `json:"text"`
	IsReply      bool   `json:"is_reply"`
	IsSmiled     bool   `json:"is_smiled"`
	IsUnsmiled   bool   `json:"is_unsmiled"`
	IsEdited     bool   `json:"is_edited"`
	RootCommID   string `json:"root_comm_id"`
	ParentCommID string `json:"parent_comm_id"`
	Depth        int    `json:"depth"`
	User         User   `json:"user"`
	Num          struct {
		Smiles   int `json:"smiles"`
		Unsmiles int `json:"unsmiles"`
		Replies  int `json:"replies"`
	} `json:"num"`
	Attachments struct {
		MentionUser []struct {
			ID           string `json:"id"`
			UserID       string `json:"user_id"`
			Nick         string `json:"nick"`
			Start        int    `json:"start_index"`
			Stop         int    `json:"stop_index"`
			OriginalNick string `json:"original_nick"`
		} `json:"mention_user"`
	} `json:"attachments"`
	ContentThumbs struct {
		URL                 string `json:"url"`
		LargeURL            string `json:"large_url"`
		X640URL             string `json:"x640_url"`
		WebpURL             string `json:"webp_url"`
		LargeWebpURL        string `json:"large_webp_url"`
		X640WebpURL         string `json:"x640_webp_url"`
		ProportionalURL     string `json:"proportional_url"`
		ProportionalWebpURL string `json:"proportional_webp_url"`
		ProportionalSize    struct {
			Width  int `json:"w"`
			Height int `json:"h"`
		} `json:"proportional_size"`
	} `json:"content_thumbs"`
}

// GetCommentPage fetches a single page of comments given a composed request.
// It is used internally by comment iteration methods and exported for advanced use cases.
func (client *Client) GetCommentPage(ctx context.Context, request compose.Request) (*Page[Comment], error) {
	content := new(struct {
		Data struct {
			Comments Page[Comment] `json:"comments"`
		}
	})

	err := client.RequestJSON(ctx, request, content)
	return &content.Data.Comments, err
}

// IterComments returns a channel that yields top-level comments on content (identified by ID).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterComments(ctx context.Context, id string) <-chan Result[*Comment] {
	return iterFrom(ctx, client, compose.Comments(id), client.GetCommentPage)
}

// GetRepliesPage fetches a single page of replies to a comment given a composed request.
// It is used internally by reply iteration methods and exported for advanced use cases.
func (client *Client) GetRepliesPage(ctx context.Context, request compose.Request) (*Page[Comment], error) {
	content := new(struct {
		Data struct {
			Replies Page[Comment] `json:"replies"`
		} `json:"data"`
	})

	err := client.RequestJSON(ctx, request, content)
	return &content.Data.Replies, err
}

// IterReplies returns a channel that yields replies to a specific comment (identified by cid on content id).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterReplies(ctx context.Context, cid, id string) <-chan Result[*Comment] {
	return iterFrom(ctx, client, compose.Replies(cid, id), client.GetRepliesPage)
}
