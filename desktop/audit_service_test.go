package main

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/evidence"
	"spreadsheet-auditor/internal/model"
)

const normalizedWorkbookPath = "<workbook>"

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), ".."))
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

func goldenJSON(t *testing.T, name string) []byte {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(repoRoot(t), "tests", "fixtures", "golden", name+".json"))
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	return payload
}

func canonicalFromCLI(t *testing.T, workbook string) []byte {
	t.Helper()
	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("cli audit workbook: %v", err)
	}
	report.WorkbookPath = normalizedWorkbookPath
	payload, err := report.CanonicalJSON()
	if err != nil {
		t.Fatalf("cli canonical json: %v", err)
	}
	return payload
}

func TestAuditServiceMatchesCLIForCombinedRiskyFixture(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()

	report, err := service.ScanWorkbook(workbook)
	if err != nil {
		t.Fatalf("service scan workbook: %v", err)
	}
	report.WorkbookPath = normalizedWorkbookPath

	serviceJSON, err := report.CanonicalJSON()
	if err != nil {
		t.Fatalf("service canonical json: %v", err)
	}

	cliJSON := canonicalFromCLI(t, workbook)
	if string(serviceJSON) != string(cliJSON) {
		t.Fatalf("service vs cli drift (%d vs %d bytes)", len(serviceJSON), len(cliJSON))
	}

	expected := goldenJSON(t, "combined_risky")
	if string(serviceJSON) != string(expected) {
		t.Fatalf("service vs golden drift (%d vs %d bytes)", len(serviceJSON), len(expected))
	}
}

func TestScanWorkbookReturnsHydratedSummaryForUI(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()

	report, err := service.ScanWorkbook(workbook)
	if err != nil {
		t.Fatalf("service scan workbook: %v", err)
	}

	expectedSummary := report.ReportSummary()
	if report.Summary.SheetCount != expectedSummary.SheetCount {
		t.Fatalf("sheet count summary drift: got %d want %d", report.Summary.SheetCount, expectedSummary.SheetCount)
	}
	if report.Summary.FormulaCellCount != expectedSummary.FormulaCellCount {
		t.Fatalf("formula count summary drift: got %d want %d", report.Summary.FormulaCellCount, expectedSummary.FormulaCellCount)
	}
	if report.Summary.IssueCount != expectedSummary.IssueCount {
		t.Fatalf("issue count summary drift: got %d want %d", report.Summary.IssueCount, expectedSummary.IssueCount)
	}
	if len(report.Summary.IssuesBySeverity) == 0 {
		t.Fatal("expected hydrated severity rollups for UI")
	}
	if len(report.Summary.IssuesByCategory) == 0 {
		t.Fatal("expected hydrated category rollups for UI")
	}
}

func TestBuildAIHandoffAlignsPromptEvidenceAndBundleJSON(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()
	options := model.PromptBundleOptions{ExcludeCells: []string{"Model!B1"}}

	handoff, err := service.BuildAIHandoff(workbook, options)
	if err != nil {
		t.Fatalf("build ai handoff: %v", err)
	}
	if handoff.AuditHash == "" {
		t.Fatal("expected audit hash on handoff payload")
	}
	if handoff.Prompt == "" {
		t.Fatal("expected prompt text on handoff payload")
	}
	if handoff.Bundle == nil {
		t.Fatal("expected structured bundle on handoff payload")
	}
	if handoff.Prompt != handoff.Bundle.Prompt {
		t.Fatal("handoff prompt drifted from bundle prompt")
	}
	if handoff.AuditHash != handoff.Bundle.EvidencePacket.AuditHash {
		t.Fatal("handoff audit hash drifted from evidence packet")
	}

	evidenceJSON, err := service.BuildEvidencePacketJSON(workbook, options)
	if err != nil {
		t.Fatalf("build evidence json: %v", err)
	}
	if handoff.EvidencePacketJSON != evidenceJSON {
		t.Fatal("handoff evidence JSON drifted from BuildEvidencePacketJSON")
	}

	bundleJSON, err := service.BuildPromptBundleJSON(workbook, options)
	if err != nil {
		t.Fatalf("build bundle json: %v", err)
	}
	if handoff.PromptBundleJSON != bundleJSON {
		t.Fatal("handoff bundle JSON drifted from BuildPromptBundleJSON")
	}
}

func TestBuildPromptBundleIncludesContractAndRedaction(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()

	bundle, err := service.BuildPromptBundle(workbook, model.PromptBundleOptions{
		ExcludeSheets: []string{"HiddenInputs"},
	})
	if err != nil {
		t.Fatalf("build prompt bundle: %v", err)
	}
	if bundle.BundleVersion != "1" {
		t.Fatalf("bundle_version = %q", bundle.BundleVersion)
	}
	if !strings.Contains(bundle.Prompt, "<<<EVIDENCE_PACKET_UNTRUSTED_BEGIN>>>") {
		t.Fatal("expected untrusted-data delimiters in prompt")
	}
	if strings.Contains(bundle.Prompt, workbook) || strings.Contains(bundle.Prompt, repoRoot(t)) {
		t.Fatal("prompt must not include absolute workbook path")
	}
	if strings.Contains(bundle.Prompt, "HiddenInputs") {
		t.Fatal("excluded sheet must not appear in prompt bundle")
	}
}

func TestSaveEvidencePacketMatchesPreparePacketOutput(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	outputPath := filepath.Join(t.TempDir(), "evidence-packet.json")
	service := NewAuditService()
	options := model.PromptBundleOptions{ExcludeCells: []string{"Model!B1"}}

	expectedJSON, err := service.BuildEvidencePacketJSON(workbook, options)
	if err != nil {
		t.Fatalf("expected json: %v", err)
	}

	if err := service.SaveEvidencePacket(workbook, outputPath, options); err != nil {
		t.Fatalf("save evidence packet: %v", err)
	}
	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read saved packet: %v", err)
	}
	if string(written) != expectedJSON {
		t.Fatal("saved evidence packet does not match prepare output")
	}
}

func TestSavePromptBundleMatchesBuildPromptBundleOutput(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	outputPath := filepath.Join(t.TempDir(), "prompt-bundle.json")
	service := NewAuditService()
	options := model.PromptBundleOptions{ExcludeCells: []string{"Model!B2"}}

	builtJSON, err := service.BuildPromptBundleJSON(workbook, options)
	if err != nil {
		t.Fatalf("built json: %v", err)
	}

	if err := service.SavePromptBundle(workbook, outputPath, options); err != nil {
		t.Fatalf("save prompt bundle: %v", err)
	}
	written, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read saved bundle: %v", err)
	}
	if string(written) != builtJSON {
		t.Fatal("saved prompt bundle does not match build output")
	}
}

func TestBuildEvidencePacketReturnsCitablePacket(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()

	packet, err := service.BuildEvidencePacket(workbook)
	if err != nil {
		t.Fatalf("build evidence packet: %v", err)
	}

	if packet.PacketVersion != "1" {
		t.Fatalf("packet version = %q", packet.PacketVersion)
	}
	if packet.AuditHash == "" {
		t.Fatal("expected audit hash")
	}
	if len(packet.Issues) == 0 {
		t.Fatal("expected audit findings in evidence packet")
	}
	if len(packet.CitationMap.IssueIDs) == 0 {
		t.Fatal("expected issue IDs in citation map")
	}
	if len(packet.CitationMap.SheetCells) == 0 {
		t.Fatal("expected sheet/cell citations in citation map")
	}
}

func TestRenderReviewPackReturnsHTML(t *testing.T) {
	workbook := fixtureWorkbook(t, "empty_workbook")
	service := NewAuditService()

	html, err := service.RenderReviewPack(workbook)
	if err != nil {
		t.Fatalf("render review pack: %v", err)
	}
	if len(html) == 0 {
		t.Fatal("expected non-empty review pack HTML")
	}
	if !strings.Contains(html, "Spreadsheet Auditor Review Pack") {
		t.Fatal("expected review pack title in HTML")
	}
}

func TestSaveReviewPackWritesFile(t *testing.T) {
	workbook := fixtureWorkbook(t, "empty_workbook")
	outputPath := filepath.Join(t.TempDir(), "review-pack.html")
	service := NewAuditService()

	if err := service.SaveReviewPack(workbook, outputPath); err != nil {
		t.Fatalf("save review pack: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read saved review pack: %v", err)
	}
	if len(content) == 0 {
		t.Fatal("expected saved review pack content")
	}
	if !strings.Contains(string(content), "Exported at:") {
		t.Fatal("expected exported-at metadata in deprecated HTML export")
	}
}

func TestSaveExportUsesPrivateFileModeAndBasenameByDefault(t *testing.T) {
	workbook := fixtureWorkbook(t, "escaping_workbook_text")
	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	service := NewAuditService()

	if err := service.SaveExport(workbook, outputPath, "csv", "2026-06-02T16:30:00Z", nil, false); err != nil {
		t.Fatalf("save csv export: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat csv: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("csv mode = %o, want 0600", info.Mode().Perm())
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if rows[1][1] != filepath.Base(report.WorkbookPath) {
		t.Fatalf("workbook_path = %q, want basename %q", rows[1][1], filepath.Base(report.WorkbookPath))
	}
}

func TestSaveExportIncludesFullPathWhenOptedIn(t *testing.T) {
	workbook := fixtureWorkbook(t, "escaping_workbook_text")
	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	service := NewAuditService()

	if err := service.SaveExport(workbook, outputPath, "csv", "2026-06-02T16:30:00Z", nil, true); err != nil {
		t.Fatalf("save csv export: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if rows[1][1] != report.WorkbookPath {
		t.Fatalf("workbook_path = %q, want full path %q", rows[1][1], report.WorkbookPath)
	}
}

func TestSaveExportWritesCSVWithExportedAt(t *testing.T) {
	workbook := fixtureWorkbook(t, "escaping_workbook_text")
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	service := NewAuditService()
	exportedAt := "2026-06-02T16:30:00Z"

	if err := service.SaveExport(workbook, outputPath, "csv", exportedAt, nil, false); err != nil {
		t.Fatalf("save csv export: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected data rows, got %d", len(rows))
	}
	if rows[1][0] != exportedAt {
		t.Fatalf("exported_at = %q", rows[1][0])
	}
}

func TestSaveExportWritesHTMLWithExportedAt(t *testing.T) {
	workbook := fixtureWorkbook(t, "empty_workbook")
	outputPath := filepath.Join(t.TempDir(), "review-pack.html")
	service := NewAuditService()
	exportedAt := "2026-06-02T16:30:00Z"

	if err := service.SaveExport(workbook, outputPath, "html", exportedAt, nil, false); err != nil {
		t.Fatalf("save html export: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read html export: %v", err)
	}
	if !strings.Contains(string(content), exportedAt) {
		t.Fatal("expected exported-at timestamp in HTML export")
	}
}

func TestSaveExportRejectsInvalidFormat(t *testing.T) {
	workbook := fixtureWorkbook(t, "empty_workbook")
	service := NewAuditService()
	err := service.SaveExport(
		workbook,
		filepath.Join(t.TempDir(), "out.bin"),
		"pdf",
		time.Now().UTC().Format(time.RFC3339),
		nil,
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected invalid format error, got %v", err)
	}
}

func TestValidateUnderstandingReportAcceptsKnownCitations(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()
	options := model.PromptBundleOptions{}

	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	report.WorkbookPath = normalizedWorkbookPath
	packet, err := evidence.BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	issueID := packet.CitationMap.IssueIDs[0]
	raw := `{
	  "workbook_purpose": [{"claim": "Review workbook risk.", "citations": ["` + issueID + `"]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`

	result, err := service.ValidateUnderstandingReport(workbook, raw, options)
	if err != nil {
		t.Fatalf("validate understanding report: %v", err)
	}
	if !result.CitationsResolved || !result.Valid {
		t.Fatalf("expected resolved citations, rejects=%#v parse=%q", result.Rejects, result.ParseError)
	}
	if result.Report == nil || len(result.Report.WorkbookPurpose) != 1 {
		t.Fatalf("expected parsed report, got %#v", result.Report)
	}
}

func TestSaveUnderstandingReportWritesValidatedJSON(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()

	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	report.WorkbookPath = normalizedWorkbookPath
	packet, err := evidence.BuildPacket(report)
	if err != nil {
		t.Fatalf("build packet: %v", err)
	}
	issueID := packet.CitationMap.IssueIDs[0]
	raw := `{
	  "workbook_purpose": [{"claim": "Review workbook risk.", "citations": ["` + issueID + `"]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`
	outputPath := filepath.Join(t.TempDir(), "verified-ai-analysis.json")

	if err := service.SaveUnderstandingReport(workbook, raw, outputPath, model.PromptBundleOptions{}); err != nil {
		t.Fatalf("save understanding report: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read saved analysis: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(content, &decoded); err != nil {
		t.Fatalf("saved analysis is not valid json: %v", err)
	}
	if _, ok := decoded["workbook_purpose"]; !ok {
		t.Fatalf("saved analysis missing workbook_purpose: %s", content)
	}
	if !strings.Contains(string(content), issueID) {
		t.Fatalf("saved analysis missing citation %q: %s", issueID, content)
	}
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("stat saved analysis: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("saved mode = %v, want 0600", got)
	}
}

func TestSaveUnderstandingReportRejectsFabricatedCitation(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()
	outputPath := filepath.Join(t.TempDir(), "fabricated.json")
	raw := `{
	  "workbook_purpose": [{"claim": "Invented.", "citations": ["fake_issue"]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`

	err := service.SaveUnderstandingReport(workbook, raw, outputPath, model.PromptBundleOptions{})
	if err == nil || !strings.Contains(err.Error(), "cited evidence did not verify") {
		t.Fatalf("expected cited evidence error, got %v", err)
	}
	if _, statErr := os.Stat(outputPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected no saved file, stat error = %v", statErr)
	}
}

func TestValidateUnderstandingReportRejectsFabricatedCitation(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()
	raw := `{
	  "workbook_purpose": [{"claim": "Invented.", "citations": ["fake_issue"]}],
	  "sheet_roles": [],
	  "key_flows": [],
	  "major_risks": [],
	  "cleanup_plan": [],
	  "owner_questions": [],
	  "confidence_notes": []
	}`

	result, err := service.ValidateUnderstandingReport(workbook, raw, model.PromptBundleOptions{})
	if err != nil {
		t.Fatalf("validate understanding report: %v", err)
	}
	if result.Valid || len(result.Rejects) != 1 || result.Rejects[0].Citation != "fake_issue" {
		t.Fatalf("expected fabricated citation rejection, got %#v", result)
	}
}

func TestValidateUnderstandingReportReturnsParseError(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	service := NewAuditService()

	result, err := service.ValidateUnderstandingReport(workbook, "{not json", model.PromptBundleOptions{})
	if err != nil {
		t.Fatalf("validate understanding report: %v", err)
	}
	if result.ParseError == "" || result.Valid {
		t.Fatalf("expected parse error, got %#v", result)
	}
}

func TestSaveExportFiltersSelectedIssueIDs(t *testing.T) {
	workbook := fixtureWorkbook(t, "combined_risky")
	report, err := audit.AuditWorkbook(workbook)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	if len(report.Issues) < 2 {
		t.Fatal("expected multiple issues")
	}
	selectedID := model.IssueID(report.Issues[0])
	outputPath := filepath.Join(t.TempDir(), "selected.csv")
	service := NewAuditService()

	if err := service.SaveExport(
		workbook,
		outputPath,
		"csv",
		"2026-06-02T16:30:00Z",
		[]string{selectedID},
		false,
	); err != nil {
		t.Fatalf("save selected export: %v", err)
	}

	file, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open csv: %v", err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header plus one selected row, got %d data rows", len(rows)-1)
	}
	if rows[1][3] != selectedID {
		t.Fatalf("issue_id = %q, want %q", rows[1][3], selectedID)
	}
}
