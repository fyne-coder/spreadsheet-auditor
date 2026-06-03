package model

import (
	"encoding/json"
	"testing"
)

func TestIssueSortingUsesLexicalCellOrder(t *testing.T) {
	issues := []Issue{
		BuildIssue("VOLATILE_FUNCTION", "m", "Order", "B2", "=TODAY()", nil),
		BuildIssue("VOLATILE_FUNCTION", "m", "Order", "B10", "=TODAY()", nil),
	}
	SortIssues(issues)
	if issues[0].Evidence.Cell != "B10" || issues[1].Evidence.Cell != "B2" {
		t.Fatalf("unexpected order: %#v", issues)
	}
}

func TestCanonicalJSONOmitsEmptyOptionalFields(t *testing.T) {
	report := &AuditReport{
		WorkbookPath:    "<workbook>",
		SupportedFormat: ".xlsx",
		Sheets: []SheetSummary{{
			Name: "Empty", State: "visible", UsedRange: "A1:A1",
		}},
		Issues: []Issue{BuildIssue(
			"BROKEN_REF_VALUE",
			"Cell contains a broken #REF! value.",
			"Empty",
			"A1",
			"",
			nil,
		)},
	}
	raw, err := report.CanonicalJSON()
	if err != nil {
		t.Fatalf("canonical json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	issues := payload["issues"].([]any)
	first := issues[0].(map[string]any)
	evidence := first["evidence"].(map[string]any)
	if _, ok := evidence["formula"]; ok {
		t.Fatalf("expected formula to be omitted, got %#v", evidence)
	}
	if _, ok := first["details"]; ok {
		t.Fatalf("expected details to be omitted")
	}
}
