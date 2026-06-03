package reviewpack

import (
	"html"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/model"
)

var fixedExportedAt = time.Date(2026, 6, 2, 12, 0, 0, 0, time.UTC)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func workbookPath(t *testing.T, name string) string {
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

func TestRenderEscapesWorkbookText(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}

	htmlOutput := RenderHTML(report, fixedExportedAt)
	if strings.Contains(htmlOutput, "<script>") {
		t.Fatal("expected escaped workbook text, found raw <script>")
	}
	if !strings.Contains(htmlOutput, html.EscapeString("Evil<script>")) {
		t.Fatal("expected escaped sheet title in HTML")
	}
	if !strings.Contains(htmlOutput, html.EscapeString("=1+<script>")) {
		t.Fatal("expected escaped formula in HTML")
	}
	if !strings.Contains(htmlOutput, html.EscapeString(report.WorkbookPath)) {
		t.Fatal("expected escaped workbook path in HTML")
	}
}

func TestRenderIncludesExpectedSections(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}

	htmlOutput := RenderHTML(report, fixedExportedAt)
	for _, section := range []string{
		"Spreadsheet Auditor Review Pack",
		"Workbook Summary",
		"Issues by Severity",
		"Issues by Category",
		"Sheet Inventory",
	} {
		if !strings.Contains(htmlOutput, section) {
			t.Fatalf("expected section %q in HTML", section)
		}
	}

	issue := report.Issues[0]
	if !strings.Contains(htmlOutput, html.EscapeString(issue.RuleID)) {
		t.Fatal("expected rule id in HTML")
	}
	if !strings.Contains(htmlOutput, html.EscapeString(issue.Evidence.Cell)) {
		t.Fatal("expected cell reference in HTML")
	}
	if !strings.Contains(htmlOutput, html.EscapeString(issue.Evidence.Formula)) {
		t.Fatal("expected formula in HTML")
	}
	if !strings.Contains(htmlOutput, html.EscapeString(issue.Remediation)) {
		t.Fatal("expected remediation in HTML")
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}

	first := RenderHTML(report, fixedExportedAt)
	second := RenderHTML(report, fixedExportedAt)
	if first != second {
		t.Fatal("expected deterministic HTML output for the same report")
	}
}

func TestRenderSeveritySpanClasses(t *testing.T) {
	report := &model.AuditReport{
		WorkbookPath:    "/tmp/example.xlsx",
		SupportedFormat: ".xlsx",
		Issues: []model.Issue{
			model.BuildIssue("HARDCODED_NUMERIC_CONSTANT", "msg", "Sheet1", "A1", "=1", nil),
		},
		Sheets: []model.SheetSummary{
			{Name: "Sheet1", State: "visible", UsedRange: "A1:A1", FormulaCells: 1},
		},
	}

	htmlOutput := RenderHTML(report, fixedExportedAt)
	if !strings.Contains(htmlOutput, `class="severity-medium"`) {
		t.Fatal("expected medium severity span class")
	}
}

func TestRenderEmptyReportSections(t *testing.T) {
	report := &model.AuditReport{
		WorkbookPath:    "/tmp/empty.xlsx",
		SupportedFormat: ".xlsx",
	}

	htmlOutput := RenderHTML(report, fixedExportedAt)
	if !strings.Contains(htmlOutput, "(none)") {
		t.Fatal("expected placeholder rows for empty report")
	}
}

func TestRenderWritesExpectedFileBytes(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "empty_workbook"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}

	outputPath := filepath.Join(t.TempDir(), "review-pack.html")
	if err := os.WriteFile(outputPath, []byte(RenderHTML(report, fixedExportedAt)), 0o644); err != nil {
		t.Fatalf("write review pack: %v", err)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected review pack file: %v", err)
	}
}
