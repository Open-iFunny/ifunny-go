package ifunny

import (
	"context"

	"github.com/gastrodon/turnpike"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

// ChatChannel represents a direct message or group chat channel. It includes the channel
// name, title, member counts, and metadata about the current user's role and membership status.
type ChatChannel struct {
	Name          string `json:"name"`
	Title         string `json:"title"`
	MembersOnline int    `json:"members_online"`
	MembersTotal  int    `json:"members_total"`

	Type      compose.ChannelType      `json:"type"`
	JoinState compose.ChannelJoinState `json:"join_state"`
	Role      compose.ChannelRole      `json:"role"`
	TouchDT   int64                    `json:"touch_dt"` // maybe when we last were online ??

	User struct {
		ID         string `json:"id"`
		Nick       string `json:"nick"`
		LastSeenAt int64  `json:"last_seen_at"`

		IsVerified bool `json:"is_verified"`
	} `json:"user"`
}

// ChatChannelPage wraps a paginated response of chat channels and a count.
type ChatChannelPage struct {
	Channels struct {
		Items  []*ChatChannel `json:"items"`
		Paging Cursor         `json:"paging"`
	} `json:"channels"`
	Num int `json:"num"`
}

func (chat *Chat) handleChannelsRaw(handle func(eventType int, channel *ChatChannel) error) EventHandler {
	return func(eventType int, kwargs map[string]any) error {
		log := chat.client.log.WithFields(logrus.Fields{"event_type": eventType, "kwargs": kwargs})

		switch eventType {
		case EVENT_JOIN, EVENT_INVITED:
			log.Trace("handle channel joined")

			for _, channelRaw := range kwargs["chats"].([]any) {
				channel := new(ChatChannel)
				if err := JSONDecode(channelRaw, channel); err != nil {
					return err
				}

				if err := handle(eventType, channel); err != nil {
					return err
				}
			}
		case EVENT_EXIT:
			log.Trace("handle channel exit")

			name := ""
			if err := JSONDecode(kwargs["chat_name"], &name); err != nil {
				return err
			}

			return handle(eventType, &ChatChannel{Name: name})
		default:
			log.Warn("no handler for this type")
		}

		return nil
	}
}

// OnChannelUpdate subscribes to channel join and invite events. The handler is called
// with the event type and channel data. Returns an unsubscribe function.
func (chat *Chat) OnChannelUpdate(handle func(eventType int, channel *ChatChannel) error) (func(), error) {
	return chat.Subscribe(compose.JoinedChannels(chat.client.Self.ID), chat.handleChannelsRaw(handle))
}

// OnChannelInvite subscribes to channel invite events. The handler is called with the
// event type and invited channel data. Returns an unsubscribe function.
func (chat *Chat) OnChannelInvite(handle func(eventType int, channel *ChatChannel) error) (func(), error) {
	return chat.Subscribe(compose.ReceiveInvite(chat.client.Self.ID), chat.handleChannelsRaw(handle))
}

// GetChannel executes a chat RPC call and unmarshals the result as a ChatChannel.
//
// The call argument is an opaque turnpike.Call — construct it with a builder
// from the [compose] package that resolves to a single channel:
//
//   - [compose.GetChannel] — by channel name
//   - [compose.GetDMChannel] — a DM channel (creates it if missing)
//
// Example (fetch a public channel by name):
//
//	channel, err := chat.GetChannel(compose.GetChannel("chat.gamers"))
//	if err != nil {
//		return err
//	}
//
// Example (open a DM with another user):
//
//	channel, err := chat.GetChannel(compose.GetDMChannel(chat.client.Self.ID, "friend-id"))
func (chat *Chat) GetChannel(call turnpike.Call) (*ChatChannel, error) {
	output := new(struct {
		Chat *ChatChannel `json:"chat"`
	})

	err := chat.Call(call, output)
	return output.Chat, err
}

// GetChannels fetches all chat channels matching the given request.
func (client *Client) GetChannels(ctx context.Context, desc compose.Request) ([]*ChatChannel, error) {
	output := new(struct {
		Data struct {
			Channels []*ChatChannel `json:"channels"`
		} `json:"data"`
	})

	err := client.RequestJSON(ctx, desc, output)
	return output.Data.Channels, err
}

// GetChannelsPage fetches a single page of chat channels given a composed request.
// It is used internally by channel iteration methods and exported for advanced use cases.
func (client *Client) GetChannelsPage(ctx context.Context, desc compose.Request) (*Page[ChatChannel], error) {
	output := new(struct {
		Data struct {
			Channels Page[ChatChannel] `json:"channels"`
		} `json:"data"`
	})
	err := client.RequestJSON(ctx, desc, output)
	return &output.Data.Channels, err
}

// IterChannelsQuery returns a channel that yields chat channels matching a search query.
// The iterator automatically fetches new pages as needed.
func (client *Client) IterChannelsQuery(ctx context.Context, query string) <-chan Result[*ChatChannel] {
	return iterFrom(ctx, client, func(limit int, page compose.Page[string]) compose.Request {
		return compose.ChatsQuery(query, limit, page)
	}, client.GetChannelsPage)
}

// IterChannelsTrending returns a channel that yields the current trending chat
// channels. The trending endpoint is a single non-paged fetch, so this is a
// one-shot iterator: it delivers every trending channel and then closes. A
// fetch error is delivered as a final Result with Err set before the channel
// closes, matching the Result/close conventions of the paginated iterators.
func (client *Client) IterChannelsTrending(ctx context.Context) <-chan Result[*ChatChannel] {
	data := make(chan Result[*ChatChannel])
	send := func(r Result[*ChatChannel]) bool {
		select {
		case data <- r:
			return true
		case <-ctx.Done():
			// Best-effort delivery of the cancellation, matching iterFrom.
			select {
			case data <- Result[*ChatChannel]{Err: ctx.Err()}:
			default:
			}
			return false
		}
	}
	go func() {
		defer close(data)
		channels, err := client.GetChannels(ctx, compose.ChatsTrending)
		if err != nil {
			send(Result[*ChatChannel]{Err: err})
			return
		}
		for _, channel := range channels {
			if !send(Result[*ChatChannel]{V: channel}) {
				return
			}
		}
	}()
	return data
}

// DMChannelName constructs a direct message channel name from the authenticated user's ID
// and one or more recipient user IDs.
func (client *Client) DMChannelName(them ...string) string {
	return compose.DMChannelName(client.Self.ID, them)
}
