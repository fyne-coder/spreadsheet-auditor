package audit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
	"spreadsheet-auditor/internal/model"
	"spreadsheet-auditor/internal/priority"
)

var supportedSuffixes = map[string]struct{}{
	".xlsx": {},
	".xlsm": {},
}

func AuditWorkbook(path string) (*model.AuditReport, error) {
	workbookPath, err := resolveWorkbookPath(path)
	if err != nil {
		return nil, err
	}

	suffix := strings.ToLower(filepath.Ext(workbookPath))
	if _, ok := supportedSuffixes[suffix]; !ok {
		return nil, fmt.Errorf(
			"unsupported workbook format %q; expected one of: .xlsm, .xlsx",
			suffix,
		)
	}

	file, err := excelize.OpenFile(workbookPath, excelize.Options{
		UnzipXMLSizeLimit: 0,
	})
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sheetNames := file.GetSheetList()
	sheetStates, err := sheetStates(file, sheetNames)
	if err != nil {
		return nil, err
	}

	var sheets []model.SheetSummary
	var issues []model.Issue

	for _, sheetName := range sheetNames {
		formulaCount, usedRange, sheetIssues, err := scanSheet(file, sheetName)
		if err != nil {
			return nil, err
		}
		issues = append(issues, sheetIssues...)

		sheets = append(sheets, model.SheetSummary{
			Name:         sheetName,
			State:        sheetStates[sheetName],
			UsedRange:    usedRange,
			FormulaCells: formulaCount,
		})
	}

	issues = append(issues, definedNameExternalRefIssues(file)...)

	model.SortIssues(issues)

	report := &model.AuditReport{
		WorkbookPath:    workbookPath,
		SupportedFormat: suffix,
		Sheets:          sheets,
		Issues:          issues,
	}
	priority.Apply(report)
	model.ApplyIssueIDs(report.Issues)
	report.Summary = report.ReportSummary()
	return report, nil
}

func resolveWorkbookPath(path string) (string, error) {
	expanded := path
	if strings.HasPrefix(path, "~/") || path == "~" {
		expanded = expandUserPath(path)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func expandUserPath(path string) string {
	if path == "~" {
		home, err := userHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := userHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
