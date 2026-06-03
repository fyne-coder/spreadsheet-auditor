package audit

import (
	"strings"

	"spreadsheet-auditor/internal/model"
)

// excelErrorCodes lists displayed/formula Excel error sentinels beyond #REF!,
// which keeps dedicated BROKEN_REF_* rules for existing parity.
var excelErrorCodes = []string{
	"#DIV/0!",
	"#VALUE!",
	"#NAME?",
	"#N/A",
	"#NUM!",
	"#NULL!",
	"#SPILL!",
	"#CALC!",
}

func displayedExcelError(value string) (code string, ok bool) {
	for _, code := range excelErrorCodes {
		if value == code {
			return code, true
		}
	}
	return "", false
}

func excelErrorValueIssue(sheet, coordinate, code string) model.Issue {
	return model.BuildIssue(
		"EXCEL_ERROR_VALUE",
		"Cell displays "+code+".",
		sheet,
		coordinate,
		"",
		map[string]any{"error_code": code},
	)
}

func excelErrorFormulaIssues(sheet, coordinate, formulaText string) []model.Issue {
	upper := strings.ToUpper(formulaText)
	var issues []model.Issue
	for _, code := range excelErrorCodes {
		if strings.Contains(upper, strings.ToUpper(code)) {
			issues = append(issues, model.BuildIssue(
				"EXCEL_ERROR_FORMULA",
				"Formula contains "+code+" error sentinel.",
				sheet,
				coordinate,
				formulaText,
				map[string]any{"error_code": code},
			))
		}
	}
	return issues
}
