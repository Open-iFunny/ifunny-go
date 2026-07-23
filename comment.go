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

// CommentsEnvelope is the response envelope for a comment feed: the page lives
// at data.comments. Hand it to [FetchPage]/[Iter] as E.
type CommentsEnvelope struct {
	Data struct {
		Comments Page[Comment] `json:"comments"`
	} `json:"data"`
}

func (e CommentsEnvelope) page() Page[Comment] { return e.Data.Comments }

// IterComments returns a channel that yields only the top-level (root) comments
// on content (identified by ID); replies are not descended into. The iterator
// automatically fetches new pages as needed. Use [Client.IterCommentsForest] to
// walk roots and all of their nested replies.
func (client *Client) IterComments(ctx context.Context, id string) <-chan Result[*Comment] {
	return Iter[CommentsEnvelope](ctx, client, compose.Comments(id))
}

// IterCommentsForest returns a channel that yields every comment on content
// (identified by ID) in depth-first order: each root comment is emitted, then
// its replies recursively (a reply and its own replies before the next sibling),
// then the next root. It stitches [Client.IterComments] and
// [Client.IterReplies] together, descending into a comment only when its
// reported reply count (Num.Replies) is non-zero to avoid empty reply requests
// on leaves.
//
// Errors and cancellation follow the same contract as the other iterators (see
// [Result]): a fetch failure anywhere in the walk is delivered as a final Result
// with Err set before the channel closes, and cancelling ctx stops the walk.
func (client *Client) IterCommentsForest(ctx context.Context, id string) <-chan Result[*Comment] {
	data := make(chan Result[*Comment])

	send := func(r Result[*Comment]) bool {
		select {
		case data <- r:
			return true
		case <-ctx.Done():
			select {
			case data <- Result[*Comment]{Err: ctx.Err()}:
			default:
			}
			return false
		}
	}

	// walk emits every comment from src, descending into the replies of any
	// comment that reports having some. It returns false as soon as delivery
	// fails (ctx cancelled) or a source yields an error, so callers unwind the
	// whole traversal.
	var walk func(src <-chan Result[*Comment]) bool
	walk = func(src <-chan Result[*Comment]) bool {
		for r := range src {
			if !send(r) {
				return false
			}
			if r.Err != nil {
				return false
			}
			if r.V.Num.Replies > 0 && !walk(client.IterReplies(ctx, id, r.V.ID)) {
				return false
			}
		}
		return true
	}

	go func() {
		defer close(data)
		walk(client.IterComments(ctx, id))
	}()

	return data
}

// RepliesEnvelope is the response envelope for a reply feed: the page lives at
// data.replies. It carries [Comment]s like [CommentsEnvelope], but this endpoint
// nests them under a different key. Hand it to [FetchPage]/[Iter] as E.
type RepliesEnvelope struct {
	Data struct {
		Replies Page[Comment] `json:"replies"`
	} `json:"data"`
}

func (e RepliesEnvelope) page() Page[Comment] { return e.Data.Replies }

// IterReplies returns a channel that yields replies to a specific comment (identified by cid on content id).
// The iterator automatically fetches new pages as needed.
func (client *Client) IterReplies(ctx context.Context, cid, id string) <-chan Result[*Comment] {
	return Iter[RepliesEnvelope](ctx, client, compose.Replies(cid, id))
}
