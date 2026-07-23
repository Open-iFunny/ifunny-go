package ifunny

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// timerTestClient spins up an httptest server that records the one request it
// receives and answers "{}", returning the wired-up client and the capture.
func timerTestClient(t *testing.T) (*Client, *http.Request, *[]byte) {
	t.Helper()

	captured := new(http.Request)
	body := new([]byte)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*captured = *r
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		*body = b
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	}))
	t.Cleanup(srv.Close)

	client, err := MakeClientBasic("dummy", Android{Version: "14"}.UserAgent(), WithAPIRoot(srv.URL))
	if err != nil {
		t.Fatalf("MakeClientBasic: %v", err)
	}
	return client, captured, body
}

// TestSetContentSchedule confirms the wire form of a schedule change: a
// urlencoded PATCH to /content/{id} carrying publish_at in unix seconds.
func TestSetContentSchedule(t *testing.T) {
	client, captured, body := timerTestClient(t)

	at := time.Unix(1893456000, 500e6) // sub-second component must be dropped
	if err := client.SetContentSchedule(context.Background(), "abc123", at); err != nil {
		t.Fatalf("SetContentSchedule: %v", err)
	}

	if captured.Method != "PATCH" {
		t.Errorf("method = %q, want PATCH", captured.Method)
	}
	if captured.URL.Path != "/content/abc123" {
		t.Errorf("path = %q, want /content/abc123", captured.URL.Path)
	}
	if ct := captured.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
		t.Errorf("content-type = %q, want application/x-www-form-urlencoded", ct)
	}

	form, err := url.ParseQuery(string(*body))
	if err != nil {
		t.Fatalf("parse body %q: %v", *body, err)
	}
	if got := form.Get("publish_at"); got != "1893456000" {
		t.Errorf("publish_at = %q, want %q", got, "1893456000")
	}
}

// TestSetContentVisibility confirms the wire form of a visibility change: a
// urlencoded PATCH to /content/{id} carrying the visibility value.
func TestSetContentVisibility(t *testing.T) {
	client, captured, body := timerTestClient(t)

	if err := client.SetContentVisibility(context.Background(), "abc123", VISIBILITY_SUBSCRIBERS); err != nil {
		t.Fatalf("SetContentVisibility: %v", err)
	}

	if captured.Method != "PATCH" {
		t.Errorf("method = %q, want PATCH", captured.Method)
	}
	if captured.URL.Path != "/content/abc123" {
		t.Errorf("path = %q, want /content/abc123", captured.URL.Path)
	}

	form, err := url.ParseQuery(string(*body))
	if err != nil {
		t.Fatalf("parse body %q: %v", *body, err)
	}
	if got := form.Get("visibility"); got != VISIBILITY_SUBSCRIBERS {
		t.Errorf("visibility = %q, want %q", got, VISIBILITY_SUBSCRIBERS)
	}
}
