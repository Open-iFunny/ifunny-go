package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

// ExploreEnvelope is the response envelope shared by every explore compilation:
// the page lives at data.value.context. Unlike the per-type feed envelopes, one
// explore endpoint serves several item types (content, users, chat channels), so
// this envelope is generic over T — pick the item type when naming it, e.g.
// ExploreEnvelope[Content]. Hand it to [FetchPage]/[Iter] as E.
type ExploreEnvelope[T Content | User | ChatChannel] struct {
	Data struct {
		Value struct {
			Context Page[T] `json:"context"`
		} `json:"value"`
	} `json:"data"`
}

func (e ExploreEnvelope[T]) page() Page[T] { return e.Data.Value.Context }

// IterExploreContent returns a channel that yields content from an explore compilation.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterExploreContent(ctx context.Context, compilation string) <-chan Result[*Content] {
	return Iter[ExploreEnvelope[Content]](ctx, client, compose.Explore(compilation))
}

// IterExploreUser returns a channel that yields users from an explore compilation.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterExploreUser(ctx context.Context, compilation string) <-chan Result[*User] {
	return Iter[ExploreEnvelope[User]](ctx, client, compose.Explore(compilation))
}

// IterExploreChatChannel returns a channel that yields chat channels from an explore compilation.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterExploreChatChannel(ctx context.Context, compilation string) <-chan Result[*ChatChannel] {
	return Iter[ExploreEnvelope[ChatChannel]](ctx, client, compose.Explore(compilation))
}
