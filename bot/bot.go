package bot

import (
	"context"

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

// MakeBot constructs a bot and authenticates its client. The bot's internal
// API calls use context.Background(); it does not currently support
// cancellation of its construction-time or per-event lookups.
func MakeBot(bearer string, ua ifunny.UserAgent) (*Bot, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetLevel(ifunny.LogLevel)
	client, err := ifunny.MakeClient(context.Background(), bearer, ua, ifunny.WithLogger(log))
	if err != nil {
		return nil, err
	}

	chat, err := client.Chat(context.Background())
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

	bot.Chat.Subscribe(context.Background(), compose.EventsIn(channel), func(eventType int, eventKW map[string]any) error {
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
