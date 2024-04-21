package ifunny

import "github.com/open-ifunny/ifunny-go/compose"

func explorePage[T Content | User | ChatChannel](client *Client, request compose.Request) (*Page[T], error) {
	content := new(struct {
		Data struct {
			Value struct {
				Context Page[T] `json:"context"`
			} `json:"value"`
		} `json:"data"`
	})

	err := client.RequestJSON(request, content)
	return &content.Data.Value.Context, err
}

func (client *Client) ExploreContentPage(requet compose.Request) (*Page[Content], error) {
	return explorePage[Content](client, requet)
}

func (client *Client) ExploreUserPage(requet compose.Request) (*Page[User], error) {
	return explorePage[User](client, requet)
}

func (client *Client) ExploreChatChannelPage(requet compose.Request) (*Page[ChatChannel], error) {
	return explorePage[ChatChannel](client, requet)
}

func iterExplore[T Content | User | ChatChannel](client *Client, compilation string) <-chan Result[*T] {
	return iterFrom(
		client,
		func(limit int, page compose.Page[string]) compose.Request {
			return compose.Explore(compilation, limit, page)
		},
		func(request compose.Request) (*Page[T], error) { return explorePage[T](client, request) },
	)
}

func (client *Client) IterExploreContent(compilation string) <-chan Result[*Content] {
	return iterExplore[Content](client, compilation)
}

func (client *Client) IterExploreUser(compilation string) <-chan Result[*User] {
	return iterExplore[User](client, compilation)
}

func (client *Client) IterExploreChatChannel(compilation string) <-chan Result[*ChatChannel] {
	return iterExplore[ChatChannel](client, compilation)
}
