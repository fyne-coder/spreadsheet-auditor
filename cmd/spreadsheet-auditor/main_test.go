package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScanWritesOutputAfterWorkbookArgument(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "report.json")
	exitCode := run([]string{
		"scan",
		fixturePath(t, "empty_workbook.xlsx"),
		"--output",
		outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestScanWritesReviewPackAfterWorkbookArgument(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "review-pack.html")
	exportedAt := "2026-06-02T12:00:00Z"
	exitCode := run([]string{
		"scan",
		fixturePath(t, "empty_workbook.xlsx"),
		"--review-pack",
		outputPath,
		"--exported-at",
		exportedAt,
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read review pack: %v", err)
	}
	html := string(content)
	if !strings.Contains(html, "Spreadsheet Auditor Review Pack") {
		t.Fatal("expected review pack title in HTML")
	}
	if !strings.Contains(html, exportedAt) {
		t.Fatal("expected exported-at metadata in review pack HTML")
	}
}

func TestScanWritesCSVExportWithExportedAt(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	exportedAt := "2026-06-02T12:00:00Z"
	exitCode := run([]string{
		"scan",
		fixturePath(t, "escaping_workbook_text.xlsx"),
		"--export-csv",
		outputPath,
		"--exported-at",
		exportedAt,
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read csv export: %v", err)
	}
	if !strings.HasPrefix(string(content), "exported_at,workbook_path,") {
		t.Fatal("expected canonical CSV header")
	}
	if !strings.Contains(string(content), exportedAt) {
		t.Fatal("expected exported-at in CSV rows")
	}
}

func TestScanCSVIncludesFullPathWhenOptedInAfterWorkbookArgument(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "review-pack.csv")
	workbookPath := fixturePath(t, "escaping_workbook_text.xlsx")
	exitCode := run([]string{
		"scan",
		workbookPath,
		"--export-csv",
		outputPath,
		"--include-full-path",
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read csv export: %v", err)
	}
	if !strings.Contains(string(content), workbookPath) {
		t.Fatal("expected opted-in full workbook path in CSV rows")
	}
}

func TestScanWritesJSONAndReviewPackTogether(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "report.json")
	htmlPath := filepath.Join(dir, "review-pack.html")
	exitCode := run([]string{
		"scan",
		fixturePath(t, "empty_workbook.xlsx"),
		"--output",
		jsonPath,
		"--review-pack",
		htmlPath,
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if _, err := os.Stat(jsonPath); err != nil {
		t.Fatalf("expected JSON output file: %v", err)
	}
	html, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("read review pack: %v", err)
	}
	if !strings.Contains(string(html), "Workbook Summary") {
		t.Fatal("expected workbook summary section in review pack")
	}
}

func TestScanWritesShortOutputAfterWorkbookArgument(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "report.json")
	exitCode := run([]string{
		"scan",
		fixturePath(t, "empty_workbook.xlsx"),
		"-o",
		outputPath,
	})
	if exitCode != 0 {
		t.Fatalf("expected exit 0, got %d", exitCode)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func fixturePath(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	return filepath.Clean(filepath.Join(
		filepath.Dir(filename),
		"..",
		"..",
		"tests",
		"fixtures",
		"workbooks",
		name,
	))
}
