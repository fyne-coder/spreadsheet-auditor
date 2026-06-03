package model

import "fmt"

// IssueID returns a stable export-selection key: rule_id|sheet|cell|message.
// Two issues that share all four fields collide; both match the same selector.
func IssueID(issue Issue) string {
	return fmt.Sprintf(
		"%s|%s|%s|%s",
		issue.RuleID,
		issue.Evidence.Sheet,
		issue.Evidence.Cell,
		issue.Message,
	)
}
