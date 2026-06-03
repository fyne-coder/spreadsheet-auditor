package audit

import (
	"strings"

	"github.com/xuri/excelize/v2"
	"spreadsheet-auditor/internal/model"
)

func definedNameExternalRefIssues(file *excelize.File) []model.Issue {
	var issues []model.Issue
	for _, definedName := range file.GetDefinedName() {
		refs := externalWorkbookRE.FindAllString(definedName.RefersTo, -1)
		if len(refs) == 0 {
			continue
		}
		scope := strings.TrimSpace(definedName.Scope)
		if scope == "" {
			scope = "Workbook"
		}
		issues = append(issues, model.BuildIssue(
			"EXTERNAL_WORKBOOK_REFERENCE",
			"Defined name references an external workbook.",
			scope,
			definedName.Name,
			definedName.RefersTo,
			map[string]any{
				"defined_name": definedName.Name,
				"refers_to":    definedName.RefersTo,
				"references":   sortedUniqueStrings(refs),
			},
		))
	}
	return issues
}
