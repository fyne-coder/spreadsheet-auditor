package priority

import (
	"sort"
	"strings"

	"spreadsheet-auditor/internal/model"
)

var structuralRuleIDs = map[string]struct{}{
	"BROKEN_REF_VALUE":    {},
	"BROKEN_REF_FORMULA":  {},
	"EXCEL_ERROR_VALUE":   {},
	"EXCEL_ERROR_FORMULA": {},
}

var riskyVolatileFunctions = map[string]struct{}{
	"CELL":        {},
	"INDIRECT":    {},
	"INFO":        {},
	"NOW":         {},
	"OFFSET":      {},
	"RAND":        {},
	"RANDBETWEEN": {},
}

var lowRiskVolatileFunctions = map[string]struct{}{
	"TODAY": {},
}

var outputLikeSheetTokens = []string{
	"output",
	"summary",
	"report",
	"dashboard",
	"total",
	"result",
}

// Apply assigns deterministic priority bands and impact factors to all issues.
func Apply(report *model.AuditReport) {
	if report == nil {
		return
	}
	states := sheetStates(report.Sheets)
	for index := range report.Issues {
		priority, factors := Assign(report.Issues[index], states[report.Issues[index].Evidence.Sheet])
		report.Issues[index].Priority = priority
		report.Issues[index].ImpactFactors = factors
	}
}

// Assign derives an analyst priority band and explainable impact factors for one issue.
func Assign(issue model.Issue, sheetState string) (string, []model.ImpactFactor) {
	factors := collectImpactFactors(issue, sheetState)
	rank := severityBaseRank(issue.Severity)

	if _, ok := structuralRuleIDs[issue.RuleID]; ok {
		rank = maxRank(rank, bandRank(model.PriorityHigh))
	}
	if issue.RuleID == "EXTERNAL_WORKBOOK_REFERENCE" {
		rank = maxRank(rank, bandRank(model.PriorityHigh))
	}
	if issue.RuleID == "FORMULA_PATTERN_ANOMALY" {
		if size := formulaClusterSize(issue); size >= 5 {
			rank = maxRank(rank, bandRank(model.PriorityHigh))
		} else if size >= 3 {
			rank = maxRank(rank, bandRank(model.PriorityMedium))
		}
	}
	if issue.RuleID == "HARDCODED_NUMERIC_CONSTANT" {
		rank = maxRank(rank, bandRank(model.PriorityMedium))
	}

	for _, factor := range factors {
		switch factor.Code {
		case "very_hidden_sheet":
			rank = bumpRank(rank, 2)
		case "hidden_sheet", "output_like_sheet", "external_workbook_dependency":
			rank = bumpRank(rank, 1)
		case "risky_volatile_functions":
			rank = bumpRank(rank, 1)
		case "formula_cluster_large":
			rank = bumpRank(rank, 2)
		case "formula_cluster_moderate":
			rank = bumpRank(rank, 1)
		case "isolated_low_risk_volatile":
			rank = bandRank(model.PriorityLow)
		}
	}

	if sheetState == "hidden" || sheetState == "veryHidden" {
		if _, ok := structuralRuleIDs[issue.RuleID]; ok {
			rank = bandRank(model.PriorityCritical)
		}
	}

	sortImpactFactors(factors)
	return rankToBand(rank), factors
}

func collectImpactFactors(issue model.Issue, sheetState string) []model.ImpactFactor {
	var factors []model.ImpactFactor

	switch sheetState {
	case "hidden":
		factors = append(factors, model.ImpactFactor{
			Code:        "hidden_sheet",
			Explanation: "Issue is on a hidden worksheet, so reviewers can miss it during a walkthrough.",
		})
	case "veryHidden":
		factors = append(factors, model.ImpactFactor{
			Code:        "very_hidden_sheet",
			Explanation: "Issue is on a very hidden worksheet, which is easy to overlook without structure review.",
		})
	}

	if outputLikeSheet(issue.Evidence.Sheet) {
		factors = append(factors, model.ImpactFactor{
			Code:        "output_like_sheet",
			Explanation: "Worksheet name suggests an output or summary surface where defects are user-visible.",
		})
	}

	if issue.RuleID == "EXTERNAL_WORKBOOK_REFERENCE" {
		factors = append(factors, model.ImpactFactor{
			Code:        "external_workbook_dependency",
			Explanation: "Formula depends on another workbook file path, version, and availability.",
		})
	}

	if issue.RuleID == "HARDCODED_NUMERIC_CONSTANT" {
		factors = append(factors, model.ImpactFactor{
			Code:        "hardcoded_numeric_constant",
			Explanation: "Control or assumption values are embedded directly in a formula instead of labeled inputs.",
		})
	}

	if size := formulaClusterSize(issue); size >= 5 {
		factors = append(factors, model.ImpactFactor{
			Code:        "formula_cluster_large",
			Explanation: "Copied formula family has at least five neighboring cells, so a local anomaly affects a wide block.",
		})
	} else if size >= 3 {
		factors = append(factors, model.ImpactFactor{
			Code:        "formula_cluster_moderate",
			Explanation: "Copied formula family spans at least three neighboring cells, increasing blast radius of a paste error.",
		})
	}

	if issue.RuleID == "VOLATILE_FUNCTION" {
		functions := volatileFunctions(issue)
		if risky, ok := riskyVolatileNames(functions); ok {
			factors = append(factors, model.ImpactFactor{
				Code:        "risky_volatile_functions",
				Explanation: "Formula uses higher-risk volatile functions: " + strings.Join(risky, ", ") + ".",
			})
		} else if isolatedLowRiskVolatile(functions) {
			factors = append(factors, model.ImpactFactor{
				Code:        "isolated_low_risk_volatile",
				Explanation: "Formula only uses TODAY(), which is volatile but usually lower priority than INDIRECT/OFFSET families.",
			})
		}
	}

	return factors
}

func sheetStates(sheets []model.SheetSummary) map[string]string {
	states := make(map[string]string, len(sheets))
	for _, sheet := range sheets {
		states[sheet.Name] = sheet.State
	}
	return states
}

func outputLikeSheet(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	for _, token := range outputLikeSheetTokens {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func formulaClusterSize(issue model.Issue) int {
	return len(stringSliceDetail(issue.Details["cluster_cells"]))
}

func volatileFunctions(issue model.Issue) []string {
	return stringSliceDetail(issue.Details["functions"])
}

func riskyVolatileNames(functions []string) ([]string, bool) {
	var risky []string
	for _, name := range functions {
		if _, ok := riskyVolatileFunctions[strings.ToUpper(name)]; ok {
			risky = append(risky, strings.ToUpper(name))
		}
	}
	if len(risky) == 0 {
		return nil, false
	}
	sort.Strings(risky)
	return risky, true
}

func isolatedLowRiskVolatile(functions []string) bool {
	if len(functions) == 0 {
		return false
	}
	for _, name := range functions {
		upper := strings.ToUpper(name)
		if _, ok := lowRiskVolatileFunctions[upper]; !ok {
			return false
		}
		if _, ok := riskyVolatileFunctions[upper]; ok {
			return false
		}
	}
	return true
}

func stringSliceDetail(value any) []string {
	switch items := value.(type) {
	case []string:
		return append([]string(nil), items...)
	case []any:
		result := make([]string, 0, len(items))
		for _, item := range items {
			if text, ok := item.(string); ok {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func severityBaseRank(severity string) int {
	switch severity {
	case "high":
		return bandRank(model.PriorityHigh)
	case "low":
		return bandRank(model.PriorityLow)
	default:
		return bandRank(model.PriorityMedium)
	}
}

func bandRank(band string) int {
	switch band {
	case model.PriorityCritical:
		return 4
	case model.PriorityHigh:
		return 3
	case model.PriorityMedium:
		return 2
	default:
		return 1
	}
}

func rankToBand(rank int) string {
	switch {
	case rank >= bandRank(model.PriorityCritical):
		return model.PriorityCritical
	case rank >= bandRank(model.PriorityHigh):
		return model.PriorityHigh
	case rank >= bandRank(model.PriorityMedium):
		return model.PriorityMedium
	default:
		return model.PriorityLow
	}
}

func maxRank(current, floor int) int {
	if current > floor {
		return current
	}
	return floor
}

func bumpRank(rank, delta int) int {
	next := rank + delta
	if next > bandRank(model.PriorityCritical) {
		return bandRank(model.PriorityCritical)
	}
	return next
}

func sortImpactFactors(factors []model.ImpactFactor) {
	sort.Slice(factors, func(i, j int) bool {
		if factors[i].Code == factors[j].Code {
			return factors[i].Explanation < factors[j].Explanation
		}
		return factors[i].Code < factors[j].Code
	})
}
