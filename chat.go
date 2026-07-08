package ifunny

import (
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
func (client *Client) Chat() (*Chat, error) {
	log := client.log.WithField("trace_id", uuid.New().String())

	log.Trace("start connect chat")
	ws, err := turnpike.NewWebsocketClient(turnpike.JSON, chatRoot, nil, nil, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Trace("join realm ifunny")
	ws.Auth = map[string]turnpike.AuthFunc{"ticket": turnpike.NewTicketAuthenticator(client.bearer)}
	hello, err := ws.JoinRealm(string(compose.URI("ifunny")), nil)
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
func (chat *Chat) Call(desc turnpike.Call, output any) error {
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": uuid.New().String(),
		"type":     "CALL",
		"uri":      desc.Procedure,
		"kwargs":   desc.ArgumentsKw,
	})

	log.Trace("exec call")
	result, err := chat.ws.Call(string(desc.Procedure), desc.Options, desc.Arguments, desc.ArgumentsKw)
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
func (chat *Chat) Publish(desc turnpike.Publish) error {
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
func (chat *Chat) Subscribe(desc turnpike.Subscribe, handle EventHandler) (func(), error) {
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": uuid.New().String(),
		"type":     "SUBSCRIBE",
		"uri":      desc.Topic,
		"options":  desc.Topic,
	})

	log.Trace("exec subscribe")
	err := chat.ws.Subscribe(string(desc.Topic), desc.Options, func(args []any, kwargs map[string]any) {
		eType := 0
		if err := JSONDecode(kwargs["type"], &eType); err != nil {
			log.Error(err)
		}

		log.WithField("event_type", eType).Trace("exec handle")
		if err := handle(eType, kwargs); err != nil {
			log.Error(err)
		}
	})

	return func() { chat.ws.Unsubscribe(string(desc.Topic)) }, err
}
