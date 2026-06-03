package reviewpack

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"spreadsheet-auditor/internal/audit"
	"spreadsheet-auditor/internal/model"
)

func TestCSVHeaderOrderIsCanonical(t *testing.T) {
	if got := CSVHeaderLine(); got != csvHeaderLine {
		t.Fatalf("header line mismatch:\n got %q\nwant %q", got, csvHeaderLine)
	}
}

func TestWriteCSVUsesFixedExportedAt(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	exportedAt := time.Date(2026, 6, 2, 15, 4, 5, 0, time.UTC)
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	if err := WriteCSV(report, outputPath, exportedAt, ExportedWorkbookPath(report.WorkbookPath, false)); err != nil {
		t.Fatalf("write csv: %v", err)
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
		t.Fatalf("expected header and data rows, got %d", len(rows))
	}
	if strings.Join(rows[0], ",") != strings.TrimSuffix(CSVHeaderLine(), "\n") {
		t.Fatalf("unexpected header row: %v", rows[0])
	}
	if rows[1][0] != "2026-06-02T15:04:05Z" {
		t.Fatalf("exported_at = %q", rows[1][0])
	}
}

func TestCSVInjectionMitigation(t *testing.T) {
	cases := map[string]string{
		"=SUM(A1)":  "'=SUM(A1)",
		"+123":      "'+123",
		"-123":      "'-123",
		"@cmd":      "'@cmd",
		"\tvalue":   "'\tvalue",
		"\rvalue":   "'\rvalue",
		"safe text": "safe text",
	}
	for input, want := range cases {
		if got := sanitizeCSVField(input); got != want {
			t.Fatalf("sanitizeCSVField(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCanonicalDetailsJSONSortedKeysWithoutTrailingNewline(t *testing.T) {
	payload, err := CanonicalDetailsJSON(map[string]any{
		"z_key": "last",
		"a_key": map[string]any{
			"nested_z": 2,
			"nested_a": 1,
		},
	})
	if err != nil {
		t.Fatalf("canonical details json: %v", err)
	}
	if strings.HasSuffix(payload, "\n") {
		t.Fatal("expected no trailing newline in details_json")
	}
	if payload != `{"a_key":{"nested_a":1,"nested_z":2},"z_key":"last"}` {
		t.Fatalf("unexpected details_json: %s", payload)
	}
}

func TestWriteExportRejectsInvalidFormat(t *testing.T) {
	report := &model.AuditReport{
		WorkbookPath:    "/tmp/example.xlsx",
		SupportedFormat: ".xlsx",
	}
	err := WriteExport(report, filepath.Join(t.TempDir(), "out.txt"), ExportOptions{
		Format:     "pdf",
		ExportedAt: fixedExportedAt,
	})
	if err == nil || !strings.Contains(err.Error(), "unsupported export format") {
		t.Fatalf("expected invalid format error, got %v", err)
	}
}

func TestFilterIssuesBySelectedIDs(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "combined_risky"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	if len(report.Issues) < 2 {
		t.Fatal("expected multiple issues in combined_risky fixture")
	}
	selectedID := model.IssueID(report.Issues[0])
	filtered := FilterIssues(report.Issues, []string{selectedID})
	if len(filtered) != 1 {
		t.Fatalf("expected one filtered issue, got %d", len(filtered))
	}
	if model.IssueID(filtered[0]) != selectedID {
		t.Fatal("filtered issue id mismatch")
	}
}

func TestWriteCSVAppliesInjectionPolicyToFormulaCells(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	if err := WriteCSV(report, outputPath, fixedExportedAt, ExportedWorkbookPath(report.WorkbookPath, false)); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.Contains(string(content), "'=1+<script>") {
		t.Fatal("expected CSV injection prefix on formula-like value")
	}
}

func TestExportedWorkbookPathDefaultsToBasename(t *testing.T) {
	got := ExportedWorkbookPath("/Users/reviewer/customer/model.xlsx", false)
	if got != "model.xlsx" {
		t.Fatalf("basename export identity = %q", got)
	}
}

func TestExportedWorkbookPathIncludesFullPathWhenOptedIn(t *testing.T) {
	full := "/Users/reviewer/customer/model.xlsx"
	if ExportedWorkbookPath(full, true) != full {
		t.Fatal("expected full workbook path when opted in")
	}
}

func TestWriteExportUsesPrivateFileMode(t *testing.T) {
	report := &model.AuditReport{
		WorkbookPath:    "/tmp/private/customer.xlsx",
		SupportedFormat: ".xlsx",
	}
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "review-pack.html")
	csvPath := filepath.Join(dir, "review-pack.csv")
	opts := ExportOptions{ExportedAt: fixedExportedAt}

	if err := WriteExport(report, htmlPath, ExportOptions{Format: FormatHTML, ExportedAt: opts.ExportedAt}); err != nil {
		t.Fatalf("write html export: %v", err)
	}
	if err := WriteExport(report, csvPath, ExportOptions{Format: FormatCSV, ExportedAt: opts.ExportedAt}); err != nil {
		t.Fatalf("write csv export: %v", err)
	}
	for _, path := range []string{htmlPath, csvPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if info.Mode().Perm() != privateExportFileMode {
			t.Fatalf("%s mode = %o, want %o", path, info.Mode().Perm(), privateExportFileMode)
		}
	}
}

func TestWriteExportDefaultsWorkbookIdentityToBasename(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	if err := WriteExport(report, outputPath, ExportOptions{
		Format:     FormatCSV,
		ExportedAt: fixedExportedAt,
	}); err != nil {
		t.Fatalf("write csv export: %v", err)
	}
	rows := readCSVRows(t, outputPath)
	if rows[1][1] != filepath.Base(report.WorkbookPath) {
		t.Fatalf("workbook_path = %q, want basename %q", rows[1][1], filepath.Base(report.WorkbookPath))
	}
}

func TestWriteExportIncludesFullPathWhenOptedIn(t *testing.T) {
	report, err := audit.AuditWorkbook(workbookPath(t, "escaping_workbook_text"))
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	if err := WriteExport(report, outputPath, ExportOptions{
		Format:          FormatCSV,
		ExportedAt:      fixedExportedAt,
		IncludeFullPath: true,
	}); err != nil {
		t.Fatalf("write csv export: %v", err)
	}
	rows := readCSVRows(t, outputPath)
	if rows[1][1] != report.WorkbookPath {
		t.Fatalf("workbook_path = %q, want full path %q", rows[1][1], report.WorkbookPath)
	}
}

func readCSVRows(t *testing.T, outputPath string) [][]string {
	t.Helper()
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
		t.Fatalf("expected header and data rows, got %d", len(rows))
	}
	return rows
}

func TestRenderHTMLIncludesExportedAtMetadata(t *testing.T) {
	report := &model.AuditReport{
		WorkbookPath:    "/tmp/example.xlsx",
		SupportedFormat: ".xlsx",
	}
	htmlOutput := RenderHTML(report, fixedExportedAt, "example.xlsx")
	if !strings.Contains(htmlOutput, "Exported at:") {
		t.Fatal("expected exported-at metadata in HTML")
	}
	if !strings.Contains(htmlOutput, "2026-06-02T12:00:00Z") {
		t.Fatal("expected RFC3339 exported-at timestamp in HTML")
	}
}
