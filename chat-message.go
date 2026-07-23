package ifunny

import (
	"context"

	"github.com/gastrodon/turnpike"
	"github.com/google/uuid"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

// Message type constants for chat events.
const (
	UNK_0        messageType = iota // Unknown type 0
	TEXT_MESSAGE                    // Text message
	UNK_2                           // Unknown type 2
	JOIN_CHANNEL                    // User joined channel
	EXIT_CHANNEL                    // User exited channel
)

type messageType int

// ChatEvent represents a single message or channel membership event in a chat.
// It includes the message text, author information, timestamp, and channel name.
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

// OnChannelEvent subscribes to messages in a channel. The handler is called for
// each message or membership event. Returns an unsubscribe function.
func (chat *Chat) OnChannelEvent(ctx context.Context, channel string, handle func(event *ChatEvent) error) (func(), error) {
	return chat.Subscribe(ctx, compose.EventsIn(channel), func(eventType int, kwargs map[string]any) error {
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

// ListMessages executes a chat RPC call and unmarshals the result as a list of messages
// with pagination cursors (prev and next).
//
// The desc argument is an opaque turnpike.Call — construct it with the
// [compose.ListMessages] builder. Prefer [Chat.IterMessages] if you want the
// whole history; ListMessages returns a single page and its cursors so the
// caller can drive pagination.
//
// Example (fetch the first page and one more):
//
//	msgs, _, next, err := chat.ListMessages(ctx, compose.ListMessages("chat.gamers", 30, compose.NONE, 0))
//	if err != nil {
//		return err
//	}
//	if next != 0 {
//		more, _, _, err := chat.ListMessages(ctx, compose.ListMessages("chat.gamers", 30, compose.NEXT, int(next)))
//		_ = more
//		_ = err
//	}
//	_ = msgs
func (chat *Chat) ListMessages(ctx context.Context, desc turnpike.Call) ([]*ChatEvent, int64, int64, error) {
	output := new(struct {
		Messages []*ChatEvent `json:"messages"`
		Prev     int64        `json:"prev"`
		Next     int64        `json:"next"`
	})

	err := chat.Call(ctx, desc, output)
	return output.Messages, output.Prev, output.Next, err
}

// IterMessages returns a channel that yields messages from a channel in chronological order.
// The iterator automatically fetches new pages as needed.
//
// The desc argument is an opaque turnpike.Call — construct it with the
// [compose.ListMessages] builder. The iterator itself carries pagination,
// so start with [compose.NoPage] rather than a cursor.
//
// Example (drain the entire history of a channel):
//
//	for result := range chat.IterMessages(ctx, compose.ListMessages("chat.gamers", 30, compose.NONE, 0)) {
//		if result.Err != nil {
//			return result.Err
//		}
//		fmt.Printf("[%s] %s\n", result.V.User.Nick, result.V.Text)
//	}
func (chat *Chat) IterMessages(ctx context.Context, desc turnpike.Call) <-chan Result[*ChatEvent] {
	data := make(chan Result[*ChatEvent])

	traceID := uuid.New().String()
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": traceID,
		"channel":  desc.ArgumentsKw["chat_name"],
	})

	send := func(r Result[*ChatEvent]) bool {
		select {
		case data <- r:
			return true
		case <-ctx.Done():
			// Best-effort delivery of the cancellation so a still-listening
			// consumer can tell cancellation from exhaustion.
			select {
			case data <- Result[*ChatEvent]{Err: ctx.Err()}:
			default:
			}
			return false
		}
	}

	go func() {
		defer close(data)
		for {
			buffer, _, next, err := chat.ListMessages(ctx, desc)
			if err != nil {
				log.Trace("failed to get a message page, exiting")
				send(Result[*ChatEvent]{Err: err})
				return
			}

			for _, event := range buffer {
				if !send(Result[*ChatEvent]{V: event}) {
					return
				}
			}

			if next == 0 || len(buffer) == 0 {
				log.Trace("reached the end of the channel, exiting")
				return
			}

			desc.ArgumentsKw["next"] = next
		}
	}()

	return data
}
