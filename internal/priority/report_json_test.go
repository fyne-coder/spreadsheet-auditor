package priority

import (
	"encoding/json"
	"testing"

	"spreadsheet-auditor/internal/model"
)

func TestCanonicalJSONIncludesPriorityAndImpactFactors(t *testing.T) {
	issue := model.BuildIssue(
		"VOLATILE_FUNCTION",
		"volatile",
		"Model",
		"B2",
		"=TODAY()",
		map[string]any{"functions": []string{"TODAY"}},
	)
	issue.Priority, issue.ImpactFactors = Assign(issue, "visible")

	report := &model.AuditReport{
		WorkbookPath:    "<workbook>",
		SupportedFormat: ".xlsx",
		Issues:          []model.Issue{issue},
	}
	raw, err := report.CanonicalJSON()
	if err != nil {
		t.Fatalf("canonical json: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	first := payload["issues"].([]any)[0].(map[string]any)
	if first["priority"] != model.PriorityLow {
		t.Fatalf("priority = %#v", first["priority"])
	}
	if len(first["impact_factors"].([]any)) == 0 {
		t.Fatal("expected impact_factors in canonical json")
	}
}
