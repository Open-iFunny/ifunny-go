package bot

import (
	"github.com/google/uuid"
	"github.com/open-ifunny/ifunny-go"
)

type filter func(ctx Context) (bool, error)
type handler func(ctx Context) error

func (bot *Bot) On(filter filter, handle handler) func() {
	handleID := uuid.New().String()
	log := bot.Log.WithField("handle_id", handleID)

	log.Trace("register on")
	bot.handleEvents[handleID] = filtHandler{filter, handle}

	return func() {
		log.Trace("delete on")
		delete(bot.handleEvents, handleID)
	}
}

func (bot *Bot) OnMessage(handle handler) func() {
	return bot.On(func(ctx Context) (bool, error) {
		if event, err := ctx.Event(); err != nil {
			return false, err
		} else {
			return event.Type == ifunny.TEXT_MESSAGE, nil
		}
	}, handle)
}

func (bot *Bot) Listen() error {
	log := bot.Log.WithField("trace_id", uuid.NewString())
	log.Trace("start event listening")

	for ctx := range bot.recvEvents {
		event, err := ctx.Event()
		if err != nil {
			log.Error(err)
			return err
		}

		eLog := log.WithField("event_id", event.ID)
		eLog.Trace("recv ctx")
		for id, filtHandle := range bot.handleEvents {
			if ok, err := filtHandle.filter(ctx); err != nil {
				eLog.Error(err)
				return err
			} else if ok {
				eLog.WithField("handle_id", id).Trace("filter ok")

				if err := filtHandle.handle(ctx); err != nil {
					eLog.WithField("handle_id", id).Error(err)
					return err
				}
			}
		}
	}

	return nil
}
