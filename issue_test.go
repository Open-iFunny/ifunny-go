package ifunny

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetNextIssueTime confirms the request hits GET /issues/next_issue_time
// and decodes the {"data": {"time_left", "time"}} envelope.
func TestGetNextIssueTime(t *testing.T) {
	var method, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"time_left":5400,"time":1893456000},"status":200}`))
	}))
	defer srv.Close()

	client, err := MakeClientBasic("dummy", Android{Version: "14"}.UserAgent(), WithAPIRoot(srv.URL))
	if err != nil {
		t.Fatalf("MakeClientBasic: %v", err)
	}

	next, err := client.GetNextIssueTime(context.Background())
	if err != nil {
		t.Fatalf("GetNextIssueTime: %v", err)
	}

	if method != "GET" {
		t.Errorf("method = %q, want GET", method)
	}
	if path != "/issues/next_issue_time" {
		t.Errorf("path = %q, want /issues/next_issue_time", path)
	}
	if next.TimeLeft != 5400 {
		t.Errorf("TimeLeft = %d, want 5400", next.TimeLeft)
	}
	if next.Time != 1893456000 {
		t.Errorf("Time = %d, want 1893456000", next.Time)
	}
}
