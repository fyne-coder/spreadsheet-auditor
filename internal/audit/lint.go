package audit

import (
	"regexp"
	"sort"
	"strings"

	"spreadsheet-auditor/internal/formula"
	"spreadsheet-auditor/internal/model"
)

var (
	wholeColumnRangeRE = regexp.MustCompile(`\$?[A-Z]{1,3}:\$?[A-Z]{1,3}`)
	externalWorkbookRE = regexp.MustCompile(`\[[^\]]+\]`)
	functionRE         = regexp.MustCompile(`\b([A-Z][A-Z0-9_.]*)\s*\(`)
)

var volatileFunctions = map[string]struct{}{
	"CELL":        {},
	"INFO":        {},
	"INDIRECT":    {},
	"NOW":         {},
	"OFFSET":      {},
	"RAND":        {},
	"RANDBETWEEN": {},
	"TODAY":       {},
}

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

	if hasWholeColumnRange(upper) {
		issues = append(issues, model.BuildIssue(
			"WHOLE_COLUMN_RANGE",
			"Formula references an entire column range, which can slow recalculation.",
			sheet,
			coordinate,
			formulaText,
			nil,
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

func volatileFunctionIssues(sheet, coordinate, formula string) []model.Issue {
	upper := strings.ToUpper(formula)
	names := map[string]struct{}{}
	for _, match := range functionRE.FindAllStringSubmatch(upper, -1) {
		name := match[1]
		if idx := strings.LastIndex(name, "."); idx >= 0 {
			name = name[idx+1:]
		}
		if _, ok := volatileFunctions[name]; ok {
			names[name] = struct{}{}
		}
	}
	if len(names) == 0 {
		return nil
	}
	volatile := make([]string, 0, len(names))
	for name := range names {
		volatile = append(volatile, name)
	}
	sort.Strings(volatile)
	return []model.Issue{model.BuildIssue(
		"VOLATILE_FUNCTION",
		"Formula uses volatile functions that can trigger expensive recalculation.",
		sheet,
		coordinate,
		formula,
		map[string]any{"functions": volatile},
	)}
}

func hasWholeColumnRange(formula string) bool {
	for _, indexes := range wholeColumnRangeRE.FindAllStringIndex(formula, -1) {
		start, end := indexes[0], indexes[1]
		if start > 0 && isFormulaIdentChar(formula[start-1]) {
			continue
		}
		if end < len(formula) && isFormulaIdentChar(formula[end]) {
			continue
		}
		return true
	}
	return false
}

func isFormulaIdentChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_'
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
