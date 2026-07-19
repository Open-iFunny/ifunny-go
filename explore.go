package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

func explorePage[T Content | User | ChatChannel](ctx context.Context, client *Client, request compose.Request) (*Page[T], error) {
	content := new(struct {
		Data struct {
			Value struct {
				Context Page[T] `json:"context"`
			} `json:"value"`
		} `json:"data"`
	})

	err := client.RequestJSON(ctx, request, content)
	return &content.Data.Value.Context, err
}

// ExploreContentPage fetches a single page of explore content given a composed request.
// It is used internally by explore iteration methods and exported for advanced use cases.
func (client *Client) ExploreContentPage(ctx context.Context, requet compose.Request) (*Page[Content], error) {
	return explorePage[Content](ctx, client, requet)
}

// ExploreUserPage fetches a single page of explore users given a composed request.
// It is used internally by explore iteration methods and exported for advanced use cases.
func (client *Client) ExploreUserPage(ctx context.Context, requet compose.Request) (*Page[User], error) {
	return explorePage[User](ctx, client, requet)
}

// ExploreChatChannelPage fetches a single page of explore chat channels given a composed request.
// It is used internally by explore iteration methods and exported for advanced use cases.
func (client *Client) ExploreChatChannelPage(ctx context.Context, requet compose.Request) (*Page[ChatChannel], error) {
	return explorePage[ChatChannel](ctx, client, requet)
}

func iterExplore[T Content | User | ChatChannel](ctx context.Context, client *Client, compilation string) <-chan Result[*T] {
	return iterFrom(
		ctx,
		client,
		compose.Explore(compilation),
		func(ctx context.Context, request compose.Request) (*Page[T], error) {
			return explorePage[T](ctx, client, request)
		},
	)
}

// IterExploreContent returns a channel that yields content from an explore compilation.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterExploreContent(ctx context.Context, compilation string) <-chan Result[*Content] {
	return iterExplore[Content](ctx, client, compilation)
}

// IterExploreUser returns a channel that yields users from an explore compilation.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterExploreUser(ctx context.Context, compilation string) <-chan Result[*User] {
	return iterExplore[User](ctx, client, compilation)
}

// IterExploreChatChannel returns a channel that yields chat channels from an explore compilation.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterExploreChatChannel(ctx context.Context, compilation string) <-chan Result[*ChatChannel] {
	return iterExplore[ChatChannel](ctx, client, compilation)
}
