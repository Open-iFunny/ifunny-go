package ifunny

// EventHandler is a callback for chat events. It receives an event type code and the
// raw event data as a map. Returns an error if event handling fails.
type EventHandler func(eventType int, kwargs map[string]any) error

// Chat event type constants.
const (
	EVENT_UNKNOWN = -1  // Unknown event type
	EVENT_JOIN    = 100 // User joined a channel
	EVENT_EXIT    = 101 // User exited a channel
	EVENT_MESSAGE = 200 // New message in channel
	EVENT_INVITED = 300 // User was invited to a channel
)
