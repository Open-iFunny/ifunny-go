package ifunny

import (
	"context"

	"github.com/google/uuid"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

// Page represents a paginated response from the iFunny API. It contains a slice
// of items of type T and pagination metadata (Cursor).
type Page[T Comment | Content | User | ChatChannel] struct {
	Items  []T    `json:"items"`
	Paging Cursor `json:"paging"`
}

// Result carries one item from an iterator or an error. Iterators close their
// channel when done; a mid-iteration failure is delivered as a final Result
// with Err set (and V zero-valued) before the channel closes.
type Result[T any] struct {
	V   T
	Err error
}

// Iterator is a convenience wrapper around a channel of Results. It holds functions
// to obtain the result channel (Iter) and to stop iteration early (Stop). Not
// currently used by the root package; it is available for future use or custom iterators.
type Iterator[T any] struct {
	Iter func() <-chan Result[T]
	Stop func()
}

func iterFrom[T Content | Comment | User | ChatChannel](ctx context.Context, client *Client, composer func(limit int, page compose.Page[string]) compose.Request, feeder func(context.Context, compose.Request) (*Page[T], error)) <-chan Result[*T] {
	page := compose.NoPage[string]()
	data := make(chan Result[*T])

	traceID := uuid.New().String()
	log := client.log.WithFields(logrus.Fields{
		"trace_id": traceID,
	})

	// send delivers r on data, but bails out if ctx is cancelled so a
	// downstream consumer that stops reading (or a cancelled request)
	// never leaks this goroutine on a blocked send.
	send := func(r Result[*T]) bool {
		select {
		case data <- r:
			return true
		case <-ctx.Done():
			return false
		}
	}

	go func() {
		defer close(data)
		for {
			log.Trace("buffering a feed page")
			items, err := feeder(ctx, composer(30, page))
			if err != nil {
				log.Trace("failed to get a feed page, exiting")
				send(Result[*T]{Err: err})
				return
			}

			for i := range items.Items {
				if !send(Result[*T]{V: &items.Items[i]}) {
					return
				}
			}

			log.Tracef("next: %s, has next: %t",
				items.Paging.Cursors.Next, items.Paging.HasNext)

			if !items.Paging.HasNext {
				log.Trace("no next page, exiting")
				return
			}

			page = compose.Next(items.Paging.Cursors.Next)
		}
	}()

	return data
}
