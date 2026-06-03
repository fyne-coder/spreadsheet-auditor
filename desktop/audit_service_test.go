package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"spreadsheet-auditor/internal/audit"
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

func TestSaveExportWritesCSVWithExportedAt(t *testing.T) {
	workbook := fixtureWorkbook(t, "escaping_workbook_text")
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	service := NewAuditService()
	exportedAt := "2026-06-02T16:30:00Z"

	if err := service.SaveExport(workbook, outputPath, "csv", exportedAt, nil); err != nil {
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

	if err := service.SaveExport(workbook, outputPath, "html", exportedAt, nil); err != nil {
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
	)
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected invalid format error, got %v", err)
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
