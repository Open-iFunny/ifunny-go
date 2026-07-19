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

// ListMessages composes a call to list messages in a channel with pagination.
// The WAMP wire wants a raw int cursor, so a Literal[int] page is unwrapped to
// its int; any other Value falls back to its String form.
func ListMessages(channel string, limit int, page Page) turnpike.Call {
	call := turnpike.Call{
		Procedure: URI("list_messages"),
		ArgumentsKw: map[string]any{
			"chat_name": channel,
			"limit":     limit,
		},
	}

	if page.Key != NONE {
		if lit, ok := page.Value.(Literal[int]); ok {
			call.ArgumentsKw[string(page.Key)] = lit.Wrapped
		} else {
			call.ArgumentsKw[string(page.Key)] = page.Value.String()
		}
	}

	return call
}
