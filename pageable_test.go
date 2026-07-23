package ifunny

import (
	"context"
	"net/http"
	"testing"

	"github.com/open-ifunny/ifunny-go/compose"
)

// TestFetchPage_SelectsEnvelope confirms the envelope type argument, not the item
// type, decides where the Page is read from: the same Content lands under
// data.content for a feed and under data.value.context for an explore compilation.
func TestFetchPage_SelectsEnvelope(t *testing.T) {
	const feedJSON = `{"data":{"content":{"items":[{"id":"feed-a"}],"paging":{"cursors":{"next":"n"},"hasNext":true}}}}`
	const explJSON = `{"data":{"value":{"context":{"items":[{"id":"expl-a"}],"paging":{"cursors":{"next":"m"},"hasNext":false}}}}}`

	t.Run("feed envelope", func(t *testing.T) {
		client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(feedJSON))
		})
		page, err := FetchPage[FeedEnvelope](context.Background(), client, compose.NamedFeed("featured").Request(compose.NoPage()))
		if err != nil {
			t.Fatalf("FetchPage: %v", err)
		}
		if len(page.Items) != 1 || page.Items[0].ID != "feed-a" {
			t.Fatalf("items: %+v", page.Items)
		}
		if page.Paging.Cursors.Next != "n" || !page.Paging.HasNext {
			t.Fatalf("paging not decoded: %+v", page.Paging)
		}
	})

	t.Run("explore envelope", func(t *testing.T) {
		client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(explJSON))
		})
		page, err := FetchPage[ExploreEnvelope[Content]](context.Background(), client, compose.Explore("content_shuffle").Request(compose.NoPage()))
		if err != nil {
			t.Fatalf("FetchPage: %v", err)
		}
		if len(page.Items) != 1 || page.Items[0].ID != "expl-a" {
			t.Fatalf("items: %+v", page.Items)
		}
	})
}

// TestIter_ThreadsPagingToken drives Iter over an explore envelope across two
// pages and asserts the server-returned cursor is sent back on the next request —
// i.e. the paging token every feed relies on rides through the generic engine via
// the envelope's page().Paging.
func TestIter_ThreadsPagingToken(t *testing.T) {
	var secondReqCursor string
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("next") {
		case "":
			// page 1: one item, hand back a cursor and promise more
			w.Write([]byte(`{"data":{"value":{"context":{"items":[{"id":"p1"}],"paging":{"cursors":{"next":"CUR1"},"hasNext":true}}}}}`))
		default:
			secondReqCursor = r.URL.Query().Get("next")
			// page 2: last item, stop
			w.Write([]byte(`{"data":{"value":{"context":{"items":[{"id":"p2"}],"paging":{"cursors":{},"hasNext":false}}}}}`))
		}
	})

	var ids []string
	for r := range Iter[ExploreEnvelope[Content]](context.Background(), client, compose.Explore("content_shuffle")) {
		if r.Err != nil {
			t.Fatalf("Iter: %v", r.Err)
		}
		ids = append(ids, r.V.ID)
	}

	if len(ids) != 2 || ids[0] != "p1" || ids[1] != "p2" {
		t.Fatalf("ids across pages: got %v, want [p1 p2]", ids)
	}
	if secondReqCursor != "CUR1" {
		t.Fatalf("page 2 cursor: got %q, want CUR1 (paging token did not thread through Iter)", secondReqCursor)
	}
}
