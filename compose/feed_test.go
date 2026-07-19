package compose

import (
	"io"
	"net/url"
	"testing"
)

func TestCollectiveRequestBodyPlacement(t *testing.T) {
	feed := Collective(0) // tail 0: verbatim cursor, still POST + body placement
	req := feed.Request(Next(IDs{"x", "y"}))

	if req.Method != "POST" {
		t.Fatalf("method: got %q, want POST", req.Method)
	}
	if req.Path != "/feeds/collective" {
		t.Fatalf("path: got %q", req.Path)
	}

	// The page token must ride in the body, not the query.
	if got := req.Query.Get("next"); got != "" {
		t.Fatalf("next leaked into query: %q", got)
	}
	if req.Query.Get("limit") != "30" {
		t.Fatalf("limit: got %q, want 30", req.Query.Get("limit"))
	}

	if ct := req.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
		t.Fatalf("content-type: got %q", ct)
	}

	if req.Body == nil {
		t.Fatal("body is nil")
	}
	raw, _ := io.ReadAll(req.Body)
	body, err := url.ParseQuery(string(raw))
	if err != nil {
		t.Fatalf("body is not urlencoded: %v", err)
	}
	if body.Get("next") != (IDs{"x", "y"}).String() {
		t.Fatalf("body next: got %q", body.Get("next"))
	}
}

func TestNamedFeedRequestQueryPlacement(t *testing.T) {
	req := NamedFeed("featured").Request(Next(Literal{Wrapped: "cursor123"}))

	if req.Method != "GET" {
		t.Fatalf("method: got %q, want GET", req.Method)
	}
	if req.Body != nil {
		t.Fatal("GET feed should have no body")
	}
	if got := req.Query.Get("next"); got != "cursor123" {
		t.Fatalf("next query param: got %q", got)
	}
}

func TestNoPageOmitsToken(t *testing.T) {
	req := NamedFeed("featured").Request(NoPage())
	if req.Query.Get("next") != "" || req.Query.Get("prev") != "" {
		t.Fatalf("NoPage should not set a token: %v", req.Query)
	}
	if req.Query.Get("limit") != "30" {
		t.Fatalf("limit: got %q", req.Query.Get("limit"))
	}
}

func TestInBodyIgnoredForGet(t *testing.T) {
	// InBody only takes effect on POST; a GET normalizes to query placement.
	feed := Feed{Path: "/feeds/x", PagerIn: InBody}
	req := feed.Request(Next(Literal{Wrapped: "tok"}))

	if req.Body != nil {
		t.Fatal("GET with InBody should not build a body")
	}
	if req.Query.Get("next") != "tok" {
		t.Fatalf("next should fall back to query: %q", req.Query.Get("next"))
	}
}

func TestCollectiveTailPaging(t *testing.T) {
	// A collective descriptor with a tail should carry a TailPager that trims
	// the server cursor to its last n IDs.
	feed := Collective(2)
	if feed.Pager == nil {
		t.Fatal("Collective(2) should set a Pager")
	}

	trimmed, err := feed.Pager((IDs{"a", "b", "c", "d"}).String())
	if err != nil {
		t.Fatalf("Pager: %v", err)
	}
	got, err := DecodeIDs(trimmed)
	if err != nil {
		t.Fatalf("DecodeIDs: %v", err)
	}
	if len(got) != 2 || got[0] != "c" || got[1] != "d" {
		t.Fatalf("tail pager: got %v, want [c d]", got)
	}
}
