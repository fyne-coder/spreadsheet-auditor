package evidence

import (
	"encoding/json"
	"testing"

	"spreadsheet-auditor/internal/model"
	"spreadsheet-auditor/internal/priority"
)

func TestBuildPacketIncludesPriorityFields(t *testing.T) {
	issue := model.BuildIssue(
		"EXTERNAL_WORKBOOK_REFERENCE",
		"external",
		"Outputs",
		"B2",
		"=[book.xlsx]Sheet1!A1",
		map[string]any{"references": []string{"[book.xlsx]"}},
	)
	band, factors := priority.Assign(issue, "visible")
	issue.Priority = band
	issue.ImpactFactors = factors

	report := &model.AuditReport{
		WorkbookPath:    "/tmp/sample.xlsx",
		SupportedFormat: ".xlsx",
		Issues:          []model.Issue{issue},
	}
	packet, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	if len(packet.Issues) != 1 {
		t.Fatalf("expected one issue, got %d", len(packet.Issues))
	}
	if packet.Issues[0].Priority != band {
		t.Fatalf("priority = %q, want %q", packet.Issues[0].Priority, band)
	}
	if len(packet.Issues[0].ImpactFactors) == 0 {
		t.Fatal("expected impact factors on evidence issue")
	}

	raw, err := packet.CanonicalJSON()
	if err != nil {
		t.Fatalf("canonical json: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	finding := decoded["audit_findings"].([]any)[0].(map[string]any)
	if finding["priority"] == nil {
		t.Fatal("expected priority in evidence packet json")
	}
}
