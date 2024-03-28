package ifunny

import (
	"github.com/gastrodon/turnpike"
	"github.com/open-ifunny/ifunny-go/compose"
)

const (
	UNK_0 messageType = iota
	TEXT_MESSAGE
	UNK_2
	JOIN_CHANNEL
	EXIT_CHANNEL
)

type messageType int

type ChatEvent struct {
	ID   string `json:"id"`
	Text string `json:"text"`

	Type   messageType `json:"type"`
	Status int         `json:"status"`
	PubAt  float64     `json:"pub_at"`

	User struct {
		ID         string `json:"user"`
		Nick       string `json:"nick"`
		IsVerified bool   `json:"is_verified"`
		LastSeenAt int64  `json:"last_seen_at"`
	} `json:"user"`

	Channel string
}

func (chat *Chat) OnChanneEvent(channel string, handle func(event *ChatEvent) error) (func(), error) {
	return chat.Subscribe(compose.EventsIn(channel), func(eventType int, kwargs map[string]interface{}) error {
		log := chat.client.log.WithField("event_type", eventType)

		if kwargs["message"] == nil {
			log.WithField("kwargs", kwargs).Warn("channel event message is nil")
		}

		message := new(ChatEvent)
		if err := JSONDecode(kwargs["message"], message); err != nil {
			return err
		}

		return handle(message)
	})
}

func (chat *Chat) ListMessages(desc turnpike.Call) ([]*ChatEvent, int64, int64, error) {
	output := new(struct {
		Messages []*ChatEvent `json:"messages"`
		Prev     int64        `json:"prev"`
		Next     int64        `json:"next"`
	})

	err := chat.Call(desc, output)
	return output.Messages, output.Prev, output.Next, err
}

func (chat *Chat) IterMessages(desc turnpike.Call) <-chan *ChatEvent {
	output := make(chan *ChatEvent)
	go func() {
		buffer, _, next, err := chat.ListMessages(desc)
		if err != nil {
			panic(err) // shrug emoji
		}

		if len(buffer) == 0 {
			close(output)
			return
		}

		for _, event := range buffer {
			output <- event
		}

		for next != 0 {
			desc.ArgumentsKw["next"] = next
			buffer, _, next, err = chat.ListMessages(desc)
			if err != nil {
				panic(err) // eye roll emoji
			}

			for _, event := range buffer {
				output <- event
			}
		}

		close(output)
	}()

	return output
}
