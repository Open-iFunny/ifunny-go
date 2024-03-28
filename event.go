package ifunny

type EventHandler func(eventType int, kwargs map[string]interface{}) error

const (
	EVENT_UNKNOWN = -1
	EVENT_JOIN    = 100
	EVENT_EXIT    = 101
	EVENT_MESSAGE = 200
	EVENT_INVITED = 300
)
