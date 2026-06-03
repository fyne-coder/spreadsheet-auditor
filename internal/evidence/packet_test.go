package evidence

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/model"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func fixtureWorkbook(t *testing.T, name string) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(repoRoot(t), "tests", "fixtures", "workbooks", name+".*"))
	if err != nil {
		t.Fatalf("glob workbook fixture: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected one workbook fixture for %s, got %v", name, matches)
	}
	return matches[0]
}

func testPacket(t *testing.T, name string) []byte {
	t.Helper()
	report, err := audit.AuditWorkbook(fixtureWorkbook(t, name))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	report.WorkbookPath = "<workbook>"
	packet, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	raw, err := packet.CanonicalJSON()
	if err != nil {
		t.Fatalf("packet json: %v", err)
	}
	return raw
}

func packetGoldenPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "tests", "fixtures", "golden", name+".packet.json")
}

func TestEvidencePacketGoldenFixtures(t *testing.T) {
	for _, name := range []string{"combined_risky", "formula_anomaly"} {
		t.Run(name, func(t *testing.T) {
			actual := testPacket(t, name)
			path := packetGoldenPath(t, name)
			if os.Getenv("UPDATE_EVIDENCE_PACKET_GOLDENS") == "1" {
				if err := os.WriteFile(path, actual, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
			}
			expected, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if string(actual) != string(expected) {
				t.Fatalf("evidence packet golden drift for %s", name)
			}
		})
	}
}

func TestEvidencePacketCanonicalJSONIsStable(t *testing.T) {
	first := testPacket(t, "combined_risky")
	second := testPacket(t, "combined_risky")
	if string(first) != string(second) {
		t.Fatal("expected stable evidence packet JSON")
	}

	var payload map[string]any
	if err := json.Unmarshal(first, &payload); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload["packet_version"] != "1" {
		t.Fatalf("packet_version = %v", payload["packet_version"])
	}
	if payload["audit_hash"] == "" {
		t.Fatal("expected audit_hash")
	}
	if !strings.Contains(string(first), "audit_findings") {
		t.Fatal("expected audit findings in packet")
	}
	if strings.Contains(string(first), fixtureWorkbook(t, "combined_risky")) ||
		strings.Contains(string(first), repoRoot(t)) {
		t.Fatal("packet must not include the absolute workbook path")
	}
	if strings.Contains(string(first), "vbaProject.bin") {
		t.Fatal("packet must not include raw VBA archive paths or bytes")
	}
}

func TestEvidencePacketAuditHashIgnoresWorkbookPath(t *testing.T) {
	report, err := audit.AuditWorkbook(fixtureWorkbook(t, "combined_risky"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	first, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build first packet: %v", err)
	}

	report.WorkbookPath = "/different/machine/customer/combined_risky.xlsx"
	second, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build second packet: %v", err)
	}

	if first.AuditHash != second.AuditHash {
		t.Fatalf("audit hash changed with path: %q vs %q", first.AuditHash, second.AuditHash)
	}
	if second.Workbook.Name != "combined_risky.xlsx" {
		t.Fatalf("expected basename only, got %q", second.Workbook.Name)
	}
}

func TestEvidencePacketDetailsUseAllowlist(t *testing.T) {
	report := &model.AuditReport{
		WorkbookPath:    "/tmp/privacy.xlsx",
		SupportedFormat: ".xlsx",
		Sheets: []model.SheetSummary{{
			Name: "Sheet1", State: "visible", UsedRange: "A1:A1", FormulaCells: 1,
		}},
		Issues: []model.Issue{model.BuildIssue(
			"HARDCODED_NUMERIC_CONSTANT",
			"Formula contains hardcoded numeric constants: 100.",
			"Sheet1",
			"A1",
			"=100",
			map[string]any{
				"constants":     []int{100},
				"sample_values": []string{"customer@example.com"},
			},
		)},
	}
	packet, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	raw, err := packet.CanonicalJSON()
	if err != nil {
		t.Fatalf("packet json: %v", err)
	}
	if strings.Contains(string(raw), "sample_values") || strings.Contains(string(raw), "customer@example.com") {
		t.Fatalf("unexpected non-allowlisted details in packet:\n%s", raw)
	}
	if !strings.Contains(string(raw), "constants") {
		t.Fatalf("expected allowlisted constants details in packet:\n%s", raw)
	}
}

func TestEvidencePacketIncludesFormulaFamilyCitation(t *testing.T) {
	raw := testPacket(t, "formula_anomaly")
	var payload struct {
		CitationMap struct {
			FormulaClusterIDs []string `json:"formula_cluster_ids"`
		} `json:"citation_map"`
		FormulaFamilies []struct {
			ClusterID   string   `json:"formula_cluster_id"`
			MemberCells []string `json:"member_cells"`
			OutlierCell string   `json:"outlier_cell"`
		} `json:"formula_families"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if len(payload.FormulaFamilies) != 1 {
		t.Fatalf("expected one formula family, got %d", len(payload.FormulaFamilies))
	}
	if payload.FormulaFamilies[0].ClusterID == "" {
		t.Fatal("expected formula cluster id")
	}
	if len(payload.CitationMap.FormulaClusterIDs) != 1 ||
		payload.CitationMap.FormulaClusterIDs[0] != payload.FormulaFamilies[0].ClusterID {
		t.Fatalf("formula cluster not in citation map: %#v", payload.CitationMap.FormulaClusterIDs)
	}
	if payload.FormulaFamilies[0].OutlierCell == "" || len(payload.FormulaFamilies[0].MemberCells) == 0 {
		t.Fatalf("expected outlier and member cells: %#v", payload.FormulaFamilies[0])
	}
}

func TestValidateUnderstandingJSONRequiresExactSectionsAndKnownCitations(t *testing.T) {
	report, err := audit.AuditWorkbook(fixtureWorkbook(t, "combined_risky"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	report.WorkbookPath = "<workbook>"
	packet, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	issueID := packet.CitationMap.IssueIDs[0]
	raw := []byte(`{
	  "workbook_purpose": [{"claim": "Review workbook risk.", "citations": ["` + issueID + `"]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`)
	if _, rejects, err := ValidateUnderstandingJSON(packet, raw); err != nil || len(rejects) != 0 {
		t.Fatalf("expected valid report, err=%v rejects=%#v", err, rejects)
	}

	missing := []byte(`{"workbook_purpose":[]}`)
	if _, _, err := ValidateUnderstandingJSON(packet, missing); err == nil ||
		!strings.Contains(err.Error(), "missing required section") {
		t.Fatalf("expected missing-section error, got %v", err)
	}

	alternate := []byte(`{
	  "schema_version": "UnderstandingReportV1",
	  "review_objective": "Help the user understand the workbook.",
	  "workbook_summary": {"interpretation": "This is a useful summary, but not the expected root schema."},
	  "deterministic_audit_state": {"issue_count": 7},
	  "major_risks": [{"risk": "Broken formulas.", "evidence": ["` + issueID + `"]}],
	  "cleanup_plan": []
	}`)
	if _, _, err := ValidateUnderstandingJSON(packet, alternate); err == nil ||
		!strings.Contains(err.Error(), "does not match UnderstandingReportV1 root schema") ||
		!strings.Contains(err.Error(), "workbook_purpose") ||
		!strings.Contains(err.Error(), "schema_version") {
		t.Fatalf("expected actionable alternate-schema error, got %v", err)
	}

	unknown := []byte(`{
	  "workbook_purpose": [{"claim": "Invented.", "citations": ["fake_issue"]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`)
	if _, rejects, err := ValidateUnderstandingJSON(packet, unknown); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	} else if len(rejects) != 1 || rejects[0].Citation != "fake_issue" {
		t.Fatalf("expected fake citation rejection, got %#v", rejects)
	}

	uncited := []byte(`{
	  "workbook_purpose": [{"claim": "Uncited.", "citations": []}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`)
	if _, rejects, err := ValidateUnderstandingJSON(packet, uncited); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	} else if len(rejects) != 1 || !strings.Contains(rejects[0].Reason, "does not include") {
		t.Fatalf("expected uncited claim rejection, got %#v", rejects)
	}
}

func TestValidateUnderstandingRejectsUnknownCitationsInEverySection(t *testing.T) {
	report, err := audit.AuditWorkbook(fixtureWorkbook(t, "combined_risky"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	report.WorkbookPath = "<workbook>"
	packet, err := BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}

	cases := []struct {
		name string
		json string
	}{
		{
			name: "workbook_purpose",
			json: `{"workbook_purpose":[{"claim":"x","citations":["fake"]}],"sheet_roles":[],"key_flows":[],"major_risks":[],"cleanup_plan":[],"owner_questions":[],"confidence_notes":[]}`,
		},
		{
			name: "sheet_roles",
			json: `{"workbook_purpose":[],"sheet_roles":[{"sheet":"Sheet1","role":"input","citations":["fake"]}],"key_flows":[],"major_risks":[],"cleanup_plan":[],"owner_questions":[],"confidence_notes":[]}`,
		},
		{
			name: "key_flows",
			json: `{"workbook_purpose":[],"sheet_roles":[],"key_flows":[{"summary":"x","citations":["fake"]}],"major_risks":[],"cleanup_plan":[],"owner_questions":[],"confidence_notes":[]}`,
		},
		{
			name: "major_risks",
			json: `{"workbook_purpose":[],"sheet_roles":[],"key_flows":[],"major_risks":[{"summary":"x","severity":"high","citations":["fake"]}],"cleanup_plan":[],"owner_questions":[],"confidence_notes":[]}`,
		},
		{
			name: "cleanup_plan",
			json: `{"workbook_purpose":[],"sheet_roles":[],"key_flows":[],"major_risks":[],"cleanup_plan":[{"action":"x","citations":["fake"]}],"owner_questions":[],"confidence_notes":[]}`,
		},
		{
			name: "owner_questions",
			json: `{"workbook_purpose":[],"sheet_roles":[],"key_flows":[],"major_risks":[],"cleanup_plan":[],"owner_questions":[{"question":"x","context_citations":["fake"]}],"confidence_notes":[]}`,
		},
		{
			name: "confidence_notes",
			json: `{"workbook_purpose":[],"sheet_roles":[],"key_flows":[],"major_risks":[],"cleanup_plan":[],"owner_questions":[],"confidence_notes":[{"note":"x","citations":["fake"]}]}`,
		},
	}

	for index, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, rejects, err := ValidateUnderstandingJSON(packet, []byte(tc.json))
			if err != nil {
				t.Fatalf("unexpected decode error: %v", err)
			}
			if len(rejects) != 1 {
				t.Fatalf("expected one reject, got %#v", rejects)
			}
			if rejects[0].Citation != "fake" || rejects[0].Index != 0 ||
				!strings.Contains(rejects[0].Field, tc.name) {
				t.Fatalf("unexpected reject for case %d: %#v", index, rejects[0])
			}
		})
	}
}

func TestValidateUnderstandingRejectsUnsupportedSection(t *testing.T) {
	packetRaw := testPacket(t, "combined_risky")
	var packet model.EvidencePacketV1
	if err := json.Unmarshal(packetRaw, &packet); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	raw := []byte(fmt.Sprintf(`{
	  "workbook_purpose": [{"claim": "x", "citations": [%q]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": [],
	  "extra": []
	}`, packet.CitationMap.IssueIDs[0]))
	if _, _, err := ValidateUnderstandingJSON(&packet, raw); err == nil ||
		!strings.Contains(err.Error(), "unsupported section") {
		t.Fatalf("expected unsupported-section error, got %v", err)
	}
}
