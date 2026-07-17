package ifunny

import (
	"context"
	"fmt"

	"github.com/gastrodon/turnpike"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

const (
	chatRoot = "wss://chat.ifunny.co/chat"
)

// Chat establishes a WebSocket connection to the iFunny chat API using the client's
// bearer token. Returns an error if authentication or connection fails.
func (client *Client) Chat(ctx context.Context) (*Chat, error) {
	log := client.log.WithField("trace_id", uuid.New().String())

	log.Trace("start connect chat")
	ws, err := turnpike.NewWebsocketClient(turnpike.JSON, chatRoot, nil, nil, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Trace("join realm ifunny")
	ws.Auth = map[string]turnpike.AuthFunc{"ticket": turnpike.NewTicketAuthenticator(client.bearer)}
	hello, err := ws.JoinRealm(ctx, string(compose.URI("ifunny")), nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &Chat{ws, client, hello}, nil
}

// Chat is a WebSocket connection to the iFunny chat API. It provides methods to
// call remote procedures (Call), publish messages (Publish), and subscribe to events
// (Subscribe). Use (*Client).Chat() to create a connection.
type Chat struct {
	ws     *turnpike.Client
	client *Client
	hello  map[string]any
}

// JSONDecode unmarshals a map or JSON-like structure into output, treating JSON
// tags on the output struct's fields as field name mappings. It is a convenience
// wrapper around mapstructure.Decode that is used throughout chat event handling.
func JSONDecode(data, output any) error {
	config := &mapstructure.DecoderConfig{TagName: "json", Result: output, WeaklyTypedInput: true}
	if decode, err := mapstructure.NewDecoder(config); err != nil {
		return err
	} else {
		return decode.Decode(data)
	}
}

// Call executes a remote procedure call (RPC) and unmarshals the result into output.
// Returns an error if the call fails or unmarshaling fails.
//
// The desc argument is an opaque turnpike.Call — construct it with a builder
// from the [compose] package. Any compose function returning turnpike.Call
// works; the ones most useful with the generic Call entry point (i.e. that
// don't already have a typed wrapper on Chat) include:
//
//   - [compose.JoinChannel], [compose.ExitChannel], [compose.HideChannel]
//   - [compose.Invite], [compose.InviteResponse], [compose.Kick]
//   - [compose.CreateChannel], [compose.NewChannel]
//
// Example (join a channel and discard the response):
//
//	var out struct{}
//	if err := chat.Call(ctx, compose.JoinChannel("chat.gamers"), &out); err != nil {
//		return err
//	}
//
// Example (invite users to a channel):
//
//	var out struct{}
//	err := chat.Call(ctx, compose.Invite("chat.gamers", []string{"12345", "67890"}), &out)
func (chat *Chat) Call(ctx context.Context, desc turnpike.Call, output any) error {
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": uuid.New().String(),
		"type":     "CALL",
		"uri":      desc.Procedure,
		"kwargs":   desc.ArgumentsKw,
	})

	log.Trace("exec call")
	result, err := chat.ws.Call(ctx, string(desc.Procedure), desc.Options, desc.Arguments, desc.ArgumentsKw)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Trace(fmt.Sprintf("call OK recv: %+v\n", result.ArgumentsKw))
	if output != nil {
		if err := JSONDecode(result.ArgumentsKw, output); err != nil {
			log.Error(err)
			return err
		}

		return nil
	}

	return nil
}

// Publish publishes a message to a topic. Returns an error if the publish fails.
//
// The desc argument is an opaque turnpike.Publish — construct it with a
// builder from the [compose] package. The only current builder is
// [compose.MessageTo].
//
// Example (send a text message to a channel):
//
//	if err := chat.Publish(ctx, compose.MessageTo("chat.gamers", "gg")); err != nil {
//		return err
//	}
func (chat *Chat) Publish(ctx context.Context, desc turnpike.Publish) error {
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": uuid.New().String(),
		"type":     "PUBLISH",
		"uri":      desc.Topic,
		"options":  desc.Topic,
	})

	log.Trace("exec publish")
	err := chat.ws.Publish(string(desc.Topic), desc.Options, desc.Arguments, desc.ArgumentsKw)
	if err != nil {
		log.Error(err)
	}

	return err
}

// Subscribe subscribes to a topic and calls handle for each event. Returns an
// unsubscribe function and an error if subscription fails.
//
// The desc argument is an opaque turnpike.Subscribe — construct it with a
// builder from the [compose] package:
//
//   - [compose.EventsIn] — messages and membership events in a channel
//   - [compose.JoinedChannels] — the authenticated user's channel join/exit events
//   - [compose.ReceiveInvite] — channel invitations sent to the authenticated user
//
// For the common cases prefer the typed wrappers [Chat.OnChannelEvent],
// [Chat.OnChannelUpdate], and [Chat.OnChannelInvite], which decode kwargs
// into strongly-typed events before calling their handler.
//
// Example (raw subscription to a channel's events):
//
//	unsubscribe, err := chat.Subscribe(ctx, compose.EventsIn("chat.gamers"), func(eventType int, kwargs map[string]any) error {
//		fmt.Printf("event %d: %+v\n", eventType, kwargs)
//		return nil
//	})
//	if err != nil {
//		return err
//	}
//	defer unsubscribe()
func (chat *Chat) Subscribe(ctx context.Context, desc turnpike.Subscribe, handle EventHandler) (func(), error) {
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": uuid.New().String(),
		"type":     "SUBSCRIBE",
		"uri":      desc.Topic,
		"options":  desc.Topic,
	})

	log.Trace("exec subscribe")
	err := chat.ws.Subscribe(ctx, string(desc.Topic), desc.Options, func(args []any, kwargs map[string]any) {
		eType := 0
		if err := JSONDecode(kwargs["type"], &eType); err != nil {
			log.Error(err)
		}

		log.WithField("event_type", eType).Trace("exec handle")
		if err := handle(eType, kwargs); err != nil {
			log.Error(err)
		}
	})

	return func() { chat.ws.Unsubscribe(ctx, string(desc.Topic)) }, err
}
