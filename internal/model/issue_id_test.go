package model

import "testing"

func TestIssueIDUsesReadablePrefixAndHashSuffix(t *testing.T) {
	issue := BuildIssue("VOLATILE_FUNCTION", "volatile formula", "Sheet1", "B2", "=NOW()", nil)
	base := IssueBaseKey(issue)
	wantPrefix := base + "|"
	got := IssueID(issue)
	if got == base {
		t.Fatalf("IssueID() = %q, expected hash suffix", got)
	}
	if len(got) <= len(wantPrefix) {
		t.Fatalf("IssueID() = %q, expected prefix %q", got, wantPrefix)
	}
	if got[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("IssueID prefix = %q, want %q", got[:len(wantPrefix)], wantPrefix)
	}
}

func TestIssueIDDistinguishesDifferentFormulaOrDetails(t *testing.T) {
	first := BuildIssue(
		"EXTERNAL_WORKBOOK_REFERENCE",
		"Formula references an external workbook.",
		"Model",
		"B4",
		"='[source.xlsx]Sheet1'!A1",
		map[string]any{"references": []string{"[source.xlsx]"}},
	)
	second := BuildIssue(
		"EXTERNAL_WORKBOOK_REFERENCE",
		"Formula references an external workbook.",
		"Model",
		"B4",
		"='[other.xlsx]Sheet1'!A1",
		map[string]any{"references": []string{"[other.xlsx]"}},
	)
	if IssueBaseKey(first) != IssueBaseKey(second) {
		t.Fatal("expected matching readable base keys")
	}
	if IssueID(first) == IssueID(second) {
		t.Fatalf("expected distinct issue IDs, both %q", IssueID(first))
	}
}

func TestIssueIDIsStableForSameIssue(t *testing.T) {
	issue := BuildIssue("VOLATILE_FUNCTION", "volatile formula", "Sheet1", "B2", "=NOW()", nil)
	if IssueID(issue) != IssueID(issue) {
		t.Fatal("expected stable issue ID")
	}
}
