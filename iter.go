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
// with Err set (and V zero-valued) before the channel closes. Cancelling the
// iterator's ctx delivers a final Result with Err = ctx.Err() on a best-effort
// basis: it is sent only if the consumer is still receiving, so a consumer
// that cancels and walks away never blocks the pager goroutine.
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

func iterFrom[T Content | Comment | User | ChatChannel](ctx context.Context, client *Client, feed compose.Feed, feeder func(context.Context, compose.Request) (*Page[T], error)) <-chan Result[*T] {
	page := compose.NoPage()
	data := make(chan Result[*T])

	traceID := uuid.New().String()
	log := client.log.WithFields(logrus.Fields{
		"trace_id": traceID,
	})

	// send delivers r on data, but bails out if ctx is cancelled so a
	// downstream consumer that stops reading (or a cancelled request)
	// never leaks this goroutine on a blocked send. On cancellation it
	// makes a best-effort (non-blocking) delivery of ctx.Err() so a
	// still-listening consumer can tell cancellation from exhaustion.
	send := func(r Result[*T]) bool {
		select {
		case data <- r:
			return true
		case <-ctx.Done():
			select {
			case data <- Result[*T]{Err: ctx.Err()}:
			default:
			}
			return false
		}
	}

	go func() {
		defer close(data)
		for {
			log.Trace("buffering a feed page")
			items, err := feeder(ctx, feed.Request(page))
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

			next := items.Paging.Cursors.Next
			if feed.Pager != nil {
				next, err = feed.Pager(next)
				if err != nil {
					log.Trace("pager failed to transform the cursor, exiting")
					send(Result[*T]{Err: err})
					return
				}
			}
			page = compose.Next(compose.Literal{Wrapped: next})
		}
	}()

	return data
}
