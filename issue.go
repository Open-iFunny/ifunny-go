package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

// NextIssueTime reports when the next batch of features (an "issue") is
// created.
type NextIssueTime struct {
	TimeLeft int64 `json:"time_left"` // seconds remaining until the drop
	Time     int64 `json:"time"`      // unix time (seconds) of the drop
}

// GetNextIssueTime fetches the time of the next featured drop.
func (client *Client) GetNextIssueTime(ctx context.Context) (*NextIssueTime, error) {
	issue := new(struct {
		Data NextIssueTime `json:"data"`
	})
	err := client.RequestJSON(ctx, compose.NextIssueTime(), issue)
	return &issue.Data, err
}
