package ifunny

import (
	"github.com/gastrodon/turnpike"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

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

type ChatChannelPage struct {
	Channels struct {
		Items  []*ChatChannel `json:"items"`
		Paging FeedCursor     `json:"paging"`
	} `json:"channels"`
	Num int `json:"num"`
}

func (chat *Chat) handleChannelsRaw(handle func(eventType int, channel *ChatChannel) error) EventHandler {
	return func(eventType int, kwargs map[string]interface{}) error {
		log := chat.client.log.WithFields(logrus.Fields{"event_type": eventType, "kwargs": kwargs})

		switch eventType {
		case EVENT_JOIN, EVENT_INVITED:
			log.Trace("handle channel joined")

			for _, channelRaw := range kwargs["chats"].([]interface{}) {
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

func (chat *Chat) OnChannelUpdate(handle func(eventType int, channel *ChatChannel) error) (func(), error) {
	return chat.Subscribe(compose.JoinedChannels(chat.client.Self.ID), chat.handleChannelsRaw(handle))
}

func (chat *Chat) OnChannelInvite(handle func(eventType int, channel *ChatChannel) error) (func(), error) {
	return chat.Subscribe(compose.ReceiveInvite(chat.client.Self.ID), chat.handleChannelsRaw(handle))
}

func (chat *Chat) GetChannel(call turnpike.Call) (*ChatChannel, error) {
	output := new(struct {
		Chat *ChatChannel `json:"chat"`
	})

	err := chat.Call(call, output)
	return output.Chat, err
}

func (client *Client) GetChannels(desc compose.Request) ([]*ChatChannel, error) {
	output := new(struct {
		Data struct {
			Channels []*ChatChannel `json:"channels"`
		} `json:"data"`
	})

	err := client.RequestJSON(desc, output)
	return output.Data.Channels, err
}

func (client *Client) GetChannelsPage(desc compose.Request) (*ChatChannelPage, error) {
	output := new(struct {
		Data ChatChannelPage `json:"data"`
	})
	err := client.RequestJSON(desc, output)
	return &output.Data, err
}

func (client *Client) IterChannels(desc compose.Request) <-chan *ChatChannel {
	output := make(chan *ChatChannel)

	go func() {
		for {
			page, err := client.GetChannelsPage(desc)
			if err != nil {
				panic(err)
			}

			for _, channel := range page.Channels.Items {
				output <- channel
			}

			if !page.Channels.Paging.HasNext {
				break
			}

			desc.Query.Set("next", page.Channels.Paging.Cursors.Next)
		}

		close(output)
	}()

	return output
}

func (client *Client) DMChannelName(them ...string) string {
	return compose.DMChannelName(client.Self.ID, them)
}
