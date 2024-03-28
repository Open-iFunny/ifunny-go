package compose

import (
	"github.com/gastrodon/turnpike"
)

/*
publish a text message to a channel
*/
func MessageTo(channel, text string) turnpike.Publish {
	return turnpike.Publish{
		Topic:     URI("chat." + channel),
		Options:   map[string]interface{}{"acknowledge": 1, "exclude_me": 1},
		Arguments: nil,
		ArgumentsKw: map[string]interface{}{
			"message_type": 1,
			"type":         200,
			"text":         text,
		},
	}
}

/*
subscribe to events happening in a channel
*/
func EventsIn(channel string) turnpike.Subscribe {
	return turnpike.Subscribe{
		Topic:   URI("chat." + channel),
		Options: nil,
	}
}

func ListMessages(channel string, limit int, page Page) turnpike.Call {
	call := turnpike.Call{
		Procedure: URI("list_messages"),
		ArgumentsKw: map[string]interface{}{
			"chat_name": channel,
			"limit":     limit,
		},
	}

	if page.Key != NONE {
		call.ArgumentsKw[string(page.Key)] = page.Value
	}

	return call
}
