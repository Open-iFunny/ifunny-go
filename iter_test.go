package ifunny

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/open-ifunny/ifunny-go/compose"
)

// feedPageJSON is a GetFeedPage-shaped response with two items and hasNext=true,
// so an iterator against it pages forever until stopped.
const feedPageJSON = `{"data":{"content":{"items":[{"id":"a"},{"id":"b"}],"paging":{"cursors":{"next":"n"},"hasNext":true}}}}`

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := MakeClientBasic("dummy", Android{Version: "14"}.UserAgent(), WithAPIRoot(srv.URL))
	if err != nil {
		t.Fatalf("MakeClientBasic: %v", err)
	}
	return client
}

// TestIterFeed_CancelClosesChannel confirms that cancelling the ctx of an
// otherwise-infinite pager tears it down: the result channel closes promptly
// instead of the pager goroutine leaking on a blocked send.
func TestIterFeed_CancelClosesChannel(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(feedPageJSON))
	})

	ctx, cancel := context.WithCancel(context.Background())
	iter := client.IterContent(ctx, compose.NamedFeed("featured"))

	// Consume one item to prove the pager is live, then cancel.
	if r := <-iter; r.Err != nil {
		t.Fatalf("first result: %v", r.Err)
	}
	cancel()

	// Drain: the channel must close promptly. Any post-cancel error result
	// must be the ctx error, not a fabricated one.
	deadline := time.After(2 * time.Second)
	for {
		select {
		case r, ok := <-iter:
			if !ok {
				return // closed: pager exited
			}
			if r.Err != nil && !errors.Is(r.Err, context.Canceled) {
				t.Fatalf("post-cancel error = %v, want context.Canceled", r.Err)
			}
		case <-deadline:
			t.Fatal("iterator did not close within 2s of cancel")
		}
	}
}

// TestIterFeed_CancelMidFetchDeliversCtxErr cancels while the pager is blocked
// on an in-flight HTTP request, and asserts a still-listening consumer receives
// a final Result carrying context.Canceled before the channel closes — i.e.
// cancellation is distinguishable from feed exhaustion.
func TestIterFeed_CancelMidFetchDeliversCtxErr(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // block until the request is cancelled
	})

	ctx, cancel := context.WithCancel(context.Background())
	iter := client.IterContent(ctx, compose.NamedFeed("featured"))

	time.AfterFunc(50*time.Millisecond, cancel)

	var sawCanceled bool
	deadline := time.After(2 * time.Second)
	for {
		select {
		case r, ok := <-iter:
			if !ok {
				if !sawCanceled {
					t.Fatal("channel closed without delivering context.Canceled")
				}
				return
			}
			if !errors.Is(r.Err, context.Canceled) {
				t.Fatalf("result error = %v, want context.Canceled", r.Err)
			}
			sawCanceled = true
		case <-deadline:
			t.Fatal("iterator did not close within 2s of cancel")
		}
	}
}

// TestIterChannelsTrending_CancelMidFetchDeliversCtxErr is the same mid-fetch
// cancellation contract for the one-shot trending iterator.
func TestIterChannelsTrending_CancelMidFetchDeliversCtxErr(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	iter := client.IterChannelsTrending(ctx)

	time.AfterFunc(50*time.Millisecond, cancel)

	var sawCanceled bool
	deadline := time.After(2 * time.Second)
	for {
		select {
		case r, ok := <-iter:
			if !ok {
				if !sawCanceled {
					t.Fatal("channel closed without delivering context.Canceled")
				}
				return
			}
			if !errors.Is(r.Err, context.Canceled) {
				t.Fatalf("result error = %v, want context.Canceled", r.Err)
			}
			sawCanceled = true
		case <-deadline:
			t.Fatal("iterator did not close within 2s of cancel")
		}
	}
}
