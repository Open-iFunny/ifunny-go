package bot

import (
	"github.com/google/uuid"
	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

type filtHandler struct {
	filter filter
	handle handler
}

type Bot struct {
	Client *ifunny.Client
	Chat   *ifunny.Chat
	Log    *logrus.Logger

	recvEvents   chan Context
	unsubEvents  map[string]func()
	handleEvents map[string]filtHandler
}

func MakeBot(bearer, userAgent string) (*Bot, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(ifunny.LogLevel)
	client, err := ifunny.MakeClientLog(bearer, userAgent, log)
	if err != nil {
		return nil, err
	}

	chat, err := client.Chat()
	if err != nil {
		return nil, err
	}

	return &Bot{
		Client:       client,
		Chat:         chat,
		Log:          log,
		recvEvents:   make(chan Context),
		unsubEvents:  make(map[string]func()),
		handleEvents: make(map[string]filtHandler, 0),
	}, nil
}

func (bot *Bot) Subscribe(channel string) {
	log := bot.Log.WithFields(logrus.Fields{"trace_id": uuid.New().String(), "channel_name": channel})
	if unsub, ok := bot.unsubEvents[channel]; ok {
		log.Warn("SubscribeChat on subscribed channel")
		unsub()
	}

	bot.Chat.Subscribe(compose.EventsIn(channel), func(eventType int, eventKW map[string]interface{}) error {
		log = log.WithFields(logrus.Fields{"event_type": eventType, "channel": channel})
		log.Trace("handle event")

		switch eventType {
		default:
			event := new(struct {
				Message ifunny.ChatEvent `json:"message"`
			})

			if err := ifunny.JSONDecode(eventKW, event); err != nil {
				log.WithField("kwargs", eventKW).Error(err)
				return err
			}

			log.Trace("push default event")
			if ctx, err := bot.makeCtx(channel, &event.Message); err != nil {
				return err
			} else {
				bot.recvEvents <- ctx
			}
		}

		return nil
	})
}

func (bot *Bot) Unsubscribe(channel string) {
	log := bot.Log.WithFields(logrus.Fields{"trace_id": uuid.New().String(), "channel_name": channel})
	if unsub, ok := bot.unsubEvents[channel]; !ok {
		log.Warn("UnsubscribeChat on not subscribed channel")
	} else {
		unsub()
	}
}
