package ifunny

import (
	"context"

	"github.com/open-ifunny/ifunny-go/compose"
)

// NextIssueTime reports when the next featured feed drop (an "issue") lands.
// This is the schedule the app's feature cadence runs on; poll it to time
// featured-feed fetches instead of guessing a fixed interval. Content.IssueAt
// records when a given post was featured.
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
