package ifunny

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

// TestIterCommentsForest_DepthFirst walks a small comment tree and asserts the
// forest iterator emits it depth-first — each root, then that root's replies
// (and their replies) before the next root — stitching the roots endpoint to the
// per-comment replies endpoint. It also confirms leaves (Num.Replies == 0) are
// never requested: the mock has no reply page for them and would 404-shaped the
// decode if they were fetched.
func TestIterCommentsForest_DepthFirst(t *testing.T) {
	// tree:
	//   A (2 replies)
	//     A1 (1 reply)
	//       A1a
	//     A2
	//   B
	roots := `{"data":{"comments":{"items":[` +
		`{"id":"A","num":{"replies":2}},` +
		`{"id":"B","num":{"replies":0}}` +
		`],"paging":{"cursors":{},"hasNext":false}}}}`
	replies := map[string]string{
		"A": `{"data":{"replies":{"items":[` +
			`{"id":"A1","num":{"replies":1}},` +
			`{"id":"A2","num":{"replies":0}}` +
			`],"paging":{"cursors":{},"hasNext":false}}}}`,
		"A1": `{"data":{"replies":{"items":[` +
			`{"id":"A1a","num":{"replies":0}}` +
			`],"paging":{"cursors":{},"hasNext":false}}}}`,
	}

	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/replies") {
			// .../comments/{commentID}/replies
			parts := strings.Split(r.URL.Path, "/")
			commentID := parts[len(parts)-2]
			body, ok := replies[commentID]
			if !ok {
				t.Errorf("unexpected replies fetch for leaf comment %q", commentID)
			}
			w.Write([]byte(body))
			return
		}
		w.Write([]byte(roots))
	})

	var got []string
	for r := range client.IterCommentsForest(context.Background(), "content1") {
		if r.Err != nil {
			t.Fatalf("forest: %v", r.Err)
		}
		got = append(got, r.V.ID)
	}

	want := []string{"A", "A1", "A1a", "A2", "B"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dfs order: got %v, want %v", got, want)
	}
}
