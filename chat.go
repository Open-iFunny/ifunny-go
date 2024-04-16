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

func (client *Client) Chat() (*Chat, error) {
	if client.scheme != BEARER {
		return nil, fmt.Errorf("cannot connect to chat without bearer token")
	}

	log := client.log.WithField("trace_id", uuid.New().String())

	log.Trace("start connect chat")
	ws, err := turnpike.NewWebsocketClient(turnpike.JSON, chatRoot, nil, nil, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	log.Trace("join realm ifunny")
	ws.Auth = map[string]turnpike.AuthFunc{"ticket": turnpike.NewTicketAuthenticator(client.token)}
	hello, err := ws.JoinRealm(string(compose.URI("ifunny")), nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &Chat{ws, client, hello}, nil
}

type Chat struct {
	ws     *turnpike.Client
	client *Client
	hello  map[string]interface{}
}

func JSONDecode(data, output interface{}) error {
	config := &mapstructure.DecoderConfig{TagName: "json", Result: output, WeaklyTypedInput: true}
	if decode, err := mapstructure.NewDecoder(config); err != nil {
		return err
	} else {
		return decode.Decode(data)
	}
}

func (chat *Chat) Call(desc turnpike.Call, output interface{}) error {
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

func (chat *Chat) Subscribe(desc turnpike.Subscribe, handle EventHandler) (func(), error) {
	log := chat.client.log.WithFields(logrus.Fields{
		"trace_id": uuid.New().String(),
		"type":     "SUBSCRIBE",
		"uri":      desc.Topic,
		"options":  desc.Topic,
	})

	log.Trace("exec subscribe")
	err := chat.ws.Subscribe(string(desc.Topic), desc.Options, func(args []interface{}, kwargs map[string]interface{}) {
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
