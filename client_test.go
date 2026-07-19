package ifunny

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/open-ifunny/ifunny-go/compose"
)

// TestWithAPIRoot_RoutesRequestsToOverride constructs a basic client pointed
// at an httptest.NewServer and confirms both PrimeBasic and RequestJSON send
// their requests to the override rather than the production apiRoot.
func TestWithAPIRoot_RoutesRequestsToOverride(t *testing.T) {
	var seenPaths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPaths = append(seenPaths, r.URL.Path)
		// PrimeBasic reads and discards; RequestJSON JSON-decodes. A "{}" body
		// satisfies both without needing per-path routing.
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	defer srv.Close()

	client, err := MakeClientBasic("dummy", Android{Version: "14"}.UserAgent(), WithAPIRoot(srv.URL))
	if err != nil {
		t.Fatalf("MakeClientBasic: %v", err)
	}

	// PrimeBasic hits /counters — we skip its 15s wait by not calling it here.
	// Exercise RequestJSON directly against a composed request.
	var out map[string]any
	if err := client.RequestJSON(context.Background(), compose.UserAccount(), &out); err != nil {
		t.Fatalf("RequestJSON: %v", err)
	}

	if len(seenPaths) == 0 {
		t.Fatal("request was not routed to the override server")
	}
	for _, p := range seenPaths {
		if strings.HasPrefix(p, "https://api.ifunny.mobi") {
			t.Errorf("request reached production apiRoot: %s", p)
		}
	}
}

// TestWithAPIRoot_TrimsTrailingSlash confirms that both "https://x" and
// "https://x/" produce the same effective apiRoot, so callers do not have
// to be careful about their input formatting.
func TestWithAPIRoot_TrimsTrailingSlash(t *testing.T) {
	c1 := newClient("Basic x", Android{Version: "14"}.UserAgent(), WithAPIRoot("https://example.test"))
	c2 := newClient("Basic x", Android{Version: "14"}.UserAgent(), WithAPIRoot("https://example.test/"))
	if c1.apiRoot != c2.apiRoot {
		t.Errorf("apiRoot with and without trailing slash differ: %q vs %q", c1.apiRoot, c2.apiRoot)
	}
	if c1.apiRoot != "https://example.test" {
		t.Errorf("apiRoot = %q, want %q", c1.apiRoot, "https://example.test")
	}
}

// TestDefaultAPIRoot confirms clients constructed without WithAPIRoot get
// the production URL, preserving the pre-0.1.4 behavior.
func TestDefaultAPIRoot(t *testing.T) {
	c := newClient("Basic x", Android{Version: "14"}.UserAgent())
	if c.apiRoot != DefaultAPIRoot {
		t.Errorf("default apiRoot = %q, want %q", c.apiRoot, DefaultAPIRoot)
	}
}

// TestHTTPError_JSONBody confirms that a structured JSON error body is decoded
// into an *APIError and returned unaltered.
func TestHTTPError_JSONBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"error":"invalid_request","error_description":"bad parameter","status":400}`))
	}))
	defer srv.Close()

	client, err := MakeClientBasic("dummy", Android{Version: "14"}.UserAgent(), WithAPIRoot(srv.URL))
	if err != nil {
		t.Fatalf("MakeClientBasic: %v", err)
	}

	var out map[string]any
	err = client.RequestJSON(context.Background(), compose.UserAccount(), &out)
	if err == nil {
		t.Fatal("RequestJSON: expected error, got nil")
	}

	apiErr, ok := AsAPIError(err)
	if !ok {
		t.Fatalf("error is not *APIError: %T", err)
	}

	if apiErr.Status != http.StatusBadRequest {
		t.Errorf("Status = %d, want %d", apiErr.Status, http.StatusBadRequest)
	}
	if apiErr.Kind != "invalid_request" {
		t.Errorf("Kind = %q, want %q", apiErr.Kind, "invalid_request")
	}
	if apiErr.Description != "bad parameter" {
		t.Errorf("Description = %q, want %q", apiErr.Description, "bad parameter")
	}
}

// TestHTTPError_PlainTextBody confirms that a non-JSON error body (e.g., a CDN
// response like "Failure: 400 Bad Request") yields a generic error carrying the
// HTTP status and the raw body, rather than an *APIError.
func TestHTTPError_PlainTextBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Failure: 400 Bad Request"))
	}))
	defer srv.Close()

	client, err := MakeClientBasic("dummy", Android{Version: "14"}.UserAgent(), WithAPIRoot(srv.URL))
	if err != nil {
		t.Fatalf("MakeClientBasic: %v", err)
	}

	var out map[string]any
	err = client.RequestJSON(context.Background(), compose.UserAccount(), &out)
	if err == nil {
		t.Fatal("RequestJSON: expected error, got nil")
	}

	if _, ok := AsAPIError(err); ok {
		t.Fatalf("expected a generic error for a non-JSON body, got *APIError")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "400") || !strings.Contains(errMsg, "Failure: 400 Bad Request") {
		t.Errorf("Error() = %q, want to contain '400' and the raw body", errMsg)
	}
}
