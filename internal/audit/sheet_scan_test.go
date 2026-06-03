package audit

import (
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
	"spreadsheet-auditor/internal/model"
)

func TestScanSheetSparseIterationSkipsBloatedDimension(t *testing.T) {
	file := excelize.NewFile()
	const sheet = "Sheet1"
	if err := file.SetCellValue(sheet, "A1", 42); err != nil {
		t.Fatalf("set cell value: %v", err)
	}
	if err := file.SetSheetDimension(sheet, "A1:XFD1048576"); err != nil {
		t.Fatalf("set sheet dimension: %v", err)
	}

	start := time.Now()
	formulaCount, usedRange, issues, err := scanSheet(file, sheet)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("scan sheet: %v", err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("scan took %v; expected sparse iteration to avoid full-rectangle walk", elapsed)
	}
	if formulaCount != 0 {
		t.Fatalf("formula count = %d, want 0", formulaCount)
	}
	if usedRange != "A1:A1" {
		t.Fatalf("used range = %q, want A1:A1", usedRange)
	}
	if len(issues) != 0 {
		t.Fatalf("issues = %d, want none for a single value cell", len(issues))
	}
}

func TestScanSheetPreservesSparseRowCoordinates(t *testing.T) {
	file := excelize.NewFile()
	const sheet = "Sheet1"
	if err := file.SetCellValue(sheet, "C10", "#VALUE!"); err != nil {
		t.Fatalf("set sparse error value: %v", err)
	}
	if err := file.SetSheetDimension(sheet, "A1:XFD1048576"); err != nil {
		t.Fatalf("set sheet dimension: %v", err)
	}

	_, usedRange, issues, err := scanSheet(file, sheet)
	if err != nil {
		t.Fatalf("scan sheet: %v", err)
	}
	if usedRange != "A1:C10" {
		t.Fatalf("used range = %q, want A1:C10", usedRange)
	}
	valueIssue := findIssue(issues, "EXCEL_ERROR_VALUE", "C10")
	if valueIssue == nil {
		t.Fatalf("expected EXCEL_ERROR_VALUE for C10, got %#v", issues)
	}
}

func TestScanSheetDetectsDisplayedAndFormulaExcelErrors(t *testing.T) {
	file := excelize.NewFile()
	const sheet = "Errors"
	if err := file.SetSheetName("Sheet1", sheet); err != nil {
		t.Fatalf("rename sheet: %v", err)
	}
	if err := file.SetCellValue(sheet, "A1", "#DIV/0!"); err != nil {
		t.Fatalf("set displayed error: %v", err)
	}
	if err := file.SetCellFormula(sheet, "B1", "=#NAME?+1"); err != nil {
		t.Fatalf("set formula error: %v", err)
	}

	_, _, issues, err := scanSheet(file, sheet)
	if err != nil {
		t.Fatalf("scan sheet: %v", err)
	}

	valueIssue := findIssue(issues, "EXCEL_ERROR_VALUE", "A1")
	if valueIssue == nil {
		t.Fatal("expected EXCEL_ERROR_VALUE for A1")
	}
	if valueIssue.Details["error_code"] != "#DIV/0!" {
		t.Fatalf("error_code = %v, want #DIV/0!", valueIssue.Details["error_code"])
	}

	formulaIssue := findIssue(issues, "EXCEL_ERROR_FORMULA", "B1")
	if formulaIssue == nil {
		t.Fatal("expected EXCEL_ERROR_FORMULA for B1")
	}
	if formulaIssue.Details["error_code"] != "#NAME?" {
		t.Fatalf("error_code = %v, want #NAME?", formulaIssue.Details["error_code"])
	}
	if formulaIssue.Evidence.Formula != "=#NAME?+1" {
		t.Fatalf("formula = %q", formulaIssue.Evidence.Formula)
	}
}

func findIssue(issues []model.Issue, ruleID, cell string) *model.Issue {
	for index := range issues {
		issue := &issues[index]
		if issue.RuleID == ruleID && issue.Evidence.Cell == cell {
			return issue
		}
	}
	return nil
}
