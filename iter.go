package ifunny

import (
	"github.com/google/uuid"
	"github.com/open-ifunny/ifunny-go/compose"
	"github.com/sirupsen/logrus"
)

type Page[T Comment | Content | User | ChatChannel] struct {
	Items  []T    `json:"items"`
	Paging Cursor `json:"paging"`
}

type Result[T any] struct {
	V   T
	Err error
}

type Iterator[T any] struct {
	Iter func() <-chan Result[T]
	Stop func()
}

func iterFrom[T Content | Comment | User | ChatChannel](client *Client, composer func(limit int, page compose.Page[string]) compose.Request, feeder func(compose.Request) (*Page[T], error)) <-chan Result[*T] {
	page := compose.NoPage[string]()
	data := make(chan Result[*T])

	traceID := uuid.New().String()
	log := client.log.WithFields(logrus.Fields{
		"trace_id": traceID,
	})

	go func() {
		defer close(data)
		for {
			log.Trace("buffering a feed page")
			items, err := feeder(composer(30, page))
			if err != nil {
				log.Trace("failed to get a feed page, exiting")
				data <- Result[*T]{Err: err}
				return
			}

			for _, v := range items.Items {
				data <- Result[*T]{V: &v}
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
