package model

import "testing"

func TestIssueIDUsesRuleSheetCellMessage(t *testing.T) {
	issue := BuildIssue("VOLATILE_FUNCTION", "volatile formula", "Sheet1", "B2", "=NOW()", nil)
	want := "VOLATILE_FUNCTION|Sheet1|B2|volatile formula"
	if got := IssueID(issue); got != want {
		t.Fatalf("IssueID() = %q, want %q", got, want)
	}
}
