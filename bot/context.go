package bot

import (
	"github.com/open-ifunny/ifunny-go"
	"github.com/open-ifunny/ifunny-go/compose"
)

type Context interface {
	Robot() *Bot
	Event() (*ifunny.ChatEvent, error)
	Caller() (*ifunny.User, error)
	Channel() (*ifunny.ChatChannel, error)
	Send(message string) error
}

type eventContext struct {
	robot       *Bot
	event       *ifunny.ChatEvent
	channelName string

	caller  *ifunny.User
	channel *ifunny.ChatChannel
}

func (ctx *eventContext) Robot() *Bot {
	return ctx.robot
}

func (ctx *eventContext) Event() (*ifunny.ChatEvent, error) {
	return ctx.event, nil
}

func (ctx *eventContext) Caller() (*ifunny.User, error) {
	if ctx.caller != nil {
		return ctx.caller, nil
	}

	var user *ifunny.User
	var err error
	if ctx.event.User.ID != "" {
		user, err = ctx.robot.Client.GetUser(compose.UserByID(ctx.event.User.ID))
	} else {
		user, err = ctx.robot.Client.GetUser(compose.UserByNick(ctx.event.User.Nick))
	}

	if err == nil {
		ctx.caller = user
	}

	return user, err
}

func (ctx *eventContext) Channel() (*ifunny.ChatChannel, error) {
	if ctx.channel != nil {
		return ctx.channel, nil
	}

	channel, err := ctx.robot.Chat.GetChannel(compose.GetChannel(ctx.channelName))
	if err == nil {
		ctx.channel = channel
	}

	return channel, nil
}

func (ctx *eventContext) Send(message string) error {
	return ctx.robot.Chat.Publish(compose.MessageTo(ctx.channelName, message))
}

func (bot *Bot) makeCtx(channel string, event *ifunny.ChatEvent) (Context, error) {
	return &eventContext{bot, event, channel, nil, nil}, nil
}
