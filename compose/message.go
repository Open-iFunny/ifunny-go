package compose

import (
	"github.com/gastrodon/turnpike"
)

// MessageTo publishes a text message to a channel.
func MessageTo(channel, text string) turnpike.Publish {
	return turnpike.Publish{
		Topic:     URI("chat." + channel),
		Options:   map[string]any{"acknowledge": 1, "exclude_me": 1},
		Arguments: nil,
		ArgumentsKw: map[string]any{
			"message_type": 1,
			"type":         200,
			"text":         text,
		},
	}
}

// EventsIn subscribes to events and messages in a channel.
func EventsIn(channel string) turnpike.Subscribe {
	return turnpike.Subscribe{
		Topic:   URI("chat." + channel),
		Options: nil,
	}
}

// ListMessages composes a call to list messages in a channel. Unlike the HTTP
// feeds, the WAMP list_messages cursor is numeric and preserves its type on the
// wire, so it does not go through [Page]/[Value]: pass the direction (NONE for
// the first page, NEXT or PREV to page) with the integer cursor from a prior
// response's Next/Prev. A NONE direction omits the cursor entirely.
func ListMessages(channel string, limit int, dir pageDirection, cursor int) turnpike.Call {
	call := turnpike.Call{
		Procedure: URI("list_messages"),
		ArgumentsKw: map[string]any{
			"chat_name": channel,
			"limit":     limit,
		},
	}

	if dir != NONE {
		call.ArgumentsKw[string(dir)] = cursor
	}

	return call
}
