package compose

// NextIssueTime composes a request for the next featured drop ("issue") time.
func NextIssueTime() Request {
	return get("/issues/next_issue_time", nil)
}
