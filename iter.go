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

// pageOf is any API response envelope that carries a [Page] of T. Each paginated
// endpoint family wraps its page in a different JSON shape (feeds under
// data.content, explore under data.value.context, subscribers under data.users,
// and so on); the envelope type is the single place that knows where its page
// lives. page returns the whole [Page] — items and the paging [Cursor] together —
// so the pagination token every feed relies on to advance rides along with the
// items rather than needing a second, envelope-specific accessor.
type pageOf[T Comment | Content | User | ChatChannel] interface {
	page() Page[T]
}

// FetchPage fetches and decodes one page from request. The response envelope is
// selected by the type argument E: FetchPage[FeedEnvelope] for a data.content
// feed, FetchPage[ExploreEnvelope[Content]] for an explore compilation,
// FetchPage[UsersEnvelope] for a subscriber list, and so on. T is inferred from
// E, so only the envelope need be named at the call site.
func FetchPage[E pageOf[T], T Comment | Content | User | ChatChannel](ctx context.Context, client *Client, request compose.Request) (*Page[T], error) {
	var env E
	if err := client.RequestJSON(ctx, request, &env); err != nil {
		return nil, err
	}
	page := env.page()
	return &page, nil
}

// Iter drives pagination over feed, decoding each page with the envelope named
// by E, and returns a channel of Results yielding *T. It is the generic engine
// behind IterContent, IterExploreContent, IterSubscribers and the rest, and the
// public escape hatch for iterating any feed descriptor against any known
// envelope — e.g. Iter[FeedEnvelope](ctx, client, compose.NamedFeed("featured")).
//
// Pagination follows the server's cursor (each page's Paging), transformed by
// feed.Pager and seeded by feed.Seed. Cancel ctx to stop; see [Result] for the
// channel-close and cancellation contract.
func Iter[E pageOf[T], T Content | Comment | User | ChatChannel](ctx context.Context, client *Client, feed compose.Feed) <-chan Result[*T] {
	page := feed.Seed
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
			items, err := FetchPage[E](ctx, client, feed.Request(page))
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
