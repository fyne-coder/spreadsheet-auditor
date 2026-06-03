package audit

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
	"spreadsheet-auditor/internal/model"
)

func TestLintFormulaIgnoresVolatileFunctionNamesInStringLiterals(t *testing.T) {
	issues := lintFormula("Model", "A1", `=IF(A1=1,"Call TODAY() before noon","")`)
	for _, issue := range issues {
		if issue.RuleID == "VOLATILE_FUNCTION" {
			t.Fatalf("unexpected volatile issue: %#v", issue)
		}
	}
}

func TestLintFormulaDetectsRandarrayDynamicArrayDetails(t *testing.T) {
	issues := lintFormula("Model", "A1", "=RANDARRAY(5,1)")
	var volatile *model.Issue
	for index := range issues {
		if issues[index].RuleID == "VOLATILE_FUNCTION" {
			volatile = &issues[index]
			break
		}
	}
	if volatile == nil {
		t.Fatalf("expected volatile issue, got %#v", issues)
	}
	if volatile.Details["dynamic_array"] != true {
		t.Fatalf("expected dynamic_array detail, got %#v", volatile.Details)
	}
	functions, ok := volatile.Details["functions"].([]string)
	if !ok || len(functions) != 1 || functions[0] != "RANDARRAY" {
		t.Fatalf("expected RANDARRAY in functions, got %#v", volatile.Details["functions"])
	}
}

func TestLintFormulaDetectsWholeRowRange(t *testing.T) {
	issues := lintFormula("Model", "A1", "=SUM(1:1)")
	if len(issues) != 1 || issues[0].RuleID != "WHOLE_COLUMN_RANGE" {
		t.Fatalf("expected whole range issue, got %#v", issues)
	}
	ranges, ok := issues[0].Details["ranges"].([]map[string]string)
	if !ok || len(ranges) != 1 || ranges[0]["kind"] != "whole_row" {
		t.Fatalf("expected whole_row detail, got %#v", issues[0].Details["ranges"])
	}
}

func TestDefinedNameExternalReferenceIssue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "defined_name_external.xlsx")
	file := excelize.NewFile()
	if err := file.SetDefinedName(&excelize.DefinedName{
		Name:     "ExternalBudget",
		RefersTo: "='[Budget Inputs.xlsx]Inputs'!$A$1:$B$10",
	}); err != nil {
		t.Fatalf("set defined name: %v", err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close workbook: %v", err)
	}

	report, err := AuditWorkbook(path)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	var found bool
	for _, issue := range report.Issues {
		if issue.RuleID != "EXTERNAL_WORKBOOK_REFERENCE" {
			continue
		}
		if issue.Evidence.Cell != "ExternalBudget" {
			continue
		}
		found = true
		if issue.Message != "Defined name references an external workbook." {
			t.Fatalf("unexpected message: %q", issue.Message)
		}
		if issue.Evidence.Formula != "='[Budget Inputs.xlsx]Inputs'!$A$1:$B$10" {
			t.Fatalf("unexpected refers_to evidence: %q", issue.Evidence.Formula)
		}
	}
	if !found {
		t.Fatalf("expected defined-name external reference issue, got %#v", report.Issues)
	}
}

func TestIssueIDCollisionResistanceInWorkbookScan(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "issue_id_collision.xlsx")
	file := excelize.NewFile()
	if err := file.SetCellFormula("Sheet1", "A1", `='[source.xlsx]Sheet1'!A1`); err != nil {
		t.Fatalf("set formula A1: %v", err)
	}
	if err := file.SetCellFormula("Sheet1", "A2", `='[other.xlsx]Sheet1'!A1`); err != nil {
		t.Fatalf("set formula A2: %v", err)
	}
	if err := file.SaveAs(path); err != nil {
		t.Fatalf("save workbook: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close workbook: %v", err)
	}

	report, err := AuditWorkbook(path)
	if err != nil {
		t.Fatalf("audit workbook: %v", err)
	}
	ids := map[string]struct{}{}
	for _, issue := range report.Issues {
		if issue.RuleID != "EXTERNAL_WORKBOOK_REFERENCE" {
			continue
		}
		id := model.IssueID(issue)
		if strings.Count(id, "|") < 4 {
			t.Fatalf("expected readable prefix plus suffix, got %q", id)
		}
		if _, ok := ids[id]; ok {
			t.Fatalf("duplicate issue id %q", id)
		}
		ids[id] = struct{}{}
	}
	if len(ids) != 2 {
		t.Fatalf("expected two distinct external-reference IDs, got %d", len(ids))
	}
}

func TestLintFormulaWholeRangeIgnoresStringLiteral(t *testing.T) {
	issues := lintFormula("Model", "A1", `=IF(A1=1,"A:A is not a range","")`)
	for _, issue := range issues {
		if issue.RuleID == "WHOLE_COLUMN_RANGE" {
			t.Fatalf("unexpected whole range issue from string literal: %#v", issue)
		}
	}
}
