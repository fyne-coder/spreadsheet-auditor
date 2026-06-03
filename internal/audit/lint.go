package audit

import (
	"regexp"
	"sort"
	"strings"

	"spreadsheet-auditor/internal/formula"
	"spreadsheet-auditor/internal/model"
)

var externalWorkbookRE = regexp.MustCompile(`\[[^\]]+\]`)

func lintFormula(sheet, coordinate, formulaText string) []model.Issue {
	var issues []model.Issue
	issues = append(issues, hardcodedNumberIssues(sheet, coordinate, formulaText)...)
	issues = append(issues, volatileFunctionIssues(sheet, coordinate, formulaText)...)

	upper := strings.ToUpper(formulaText)
	if strings.Contains(upper, "#REF!") {
		issues = append(issues, model.BuildIssue(
			"BROKEN_REF_FORMULA",
			"Formula contains a broken #REF! reference.",
			sheet,
			coordinate,
			formulaText,
			nil,
		))
	}

	issues = append(issues, excelErrorFormulaIssues(sheet, coordinate, formulaText)...)

	if rangeRefs := formula.WholeRangeReferences(formulaText); len(rangeRefs) > 0 {
		issues = append(issues, model.BuildIssue(
			"WHOLE_COLUMN_RANGE",
			"Formula references an entire column range, which can slow recalculation.",
			sheet,
			coordinate,
			formulaText,
			wholeRangeDetails(rangeRefs),
		))
	}

	refs := externalWorkbookRE.FindAllString(formulaText, -1)
	if len(refs) > 0 {
		unique := sortedUniqueStrings(refs)
		issues = append(issues, model.BuildIssue(
			"EXTERNAL_WORKBOOK_REFERENCE",
			"Formula references an external workbook.",
			sheet,
			coordinate,
			formulaText,
			map[string]any{"references": unique},
		))
	}

	return issues
}

func hardcodedNumberIssues(sheet, coordinate, formulaText string) []model.Issue {
	constants := formula.NumberOperands(formula.Tokenize(formulaText))
	if len(constants) == 0 {
		return nil
	}
	return []model.Issue{model.BuildIssue(
		"HARDCODED_NUMERIC_CONSTANT",
		"Formula contains hardcoded numeric constants.",
		sheet,
		coordinate,
		formulaText,
		map[string]any{"constants": constants},
	)}
}

func volatileFunctionIssues(sheet, coordinate, formulaText string) []model.Issue {
	result := formula.VolatileFunctions(formulaText)
	if len(result.Functions) == 0 {
		return nil
	}
	details := map[string]any{"functions": result.Functions}
	if result.DynamicArray {
		details["dynamic_array"] = true
		details["dynamic_array_functions"] = result.DynamicNames
	}
	return []model.Issue{model.BuildIssue(
		"VOLATILE_FUNCTION",
		"Formula uses volatile functions that can trigger expensive recalculation.",
		sheet,
		coordinate,
		formulaText,
		details,
	)}
}

func wholeRangeDetails(refs []formula.RangeReference) map[string]any {
	if !needsExpandedRangeDetails(refs) {
		return nil
	}
	return map[string]any{"ranges": rangeReferencesToMaps(refs)}
}

func needsExpandedRangeDetails(refs []formula.RangeReference) bool {
	for _, ref := range refs {
		if ref.Kind != "whole_column" {
			return true
		}
	}
	return false
}

func rangeReferencesToMaps(refs []formula.RangeReference) []map[string]string {
	out := make([]map[string]string, 0, len(refs))
	for _, ref := range refs {
		out = append(out, map[string]string{
			"kind":      ref.Kind,
			"reference": ref.Reference,
		})
	}
	return out
}

func sortedUniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		seen[value] = struct{}{}
	}
	unique := make([]string, 0, len(seen))
	for value := range seen {
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}
