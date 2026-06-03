package promptpack

import (
	"sort"
	"strings"

	"spreadsheet-auditor/internal/model"
)

func applyExclusions(packet *model.EvidencePacketV1, options model.PromptBundleOptions) {
	excludedSheets := stringSet(options.ExcludeSheets)
	excludedCells := stringSet(options.ExcludeCells)

	if len(excludedSheets) == 0 && len(excludedCells) == 0 {
		return
	}

	filteredSheets := make([]model.EvidenceSheet, 0, len(packet.Sheets))
	for _, sheet := range packet.Sheets {
		if _, excluded := excludedSheets[sheet.Name]; excluded {
			continue
		}
		filteredSheets = append(filteredSheets, sheet)
	}
	packet.Sheets = filteredSheets

	filteredIssues := make([]model.EvidenceIssue, 0, len(packet.Issues))
	for _, issue := range packet.Issues {
		if _, excluded := excludedSheets[issue.Sheet]; excluded {
			continue
		}
		if issue.Sheet != "" && issue.Cell != "" {
			if isExcludedCell(issue.Sheet, issue.Cell, excludedCells) {
				continue
			}
		}
		issue.Details = filterDetails(issue.Sheet, issue.Details, excludedCells)
		filteredIssues = append(filteredIssues, issue)
	}
	packet.Issues = filteredIssues

	filteredFamilies := make([]model.FormulaFamily, 0, len(packet.FormulaFamilies))
	for _, family := range packet.FormulaFamilies {
		if _, excluded := excludedSheets[family.Sheet]; excluded {
			continue
		}
		if isExcludedCell(family.Sheet, family.OutlierCell, excludedCells) {
			continue
		}
		memberCells := make([]string, 0, len(family.MemberCells))
		for _, cell := range family.MemberCells {
			if isExcludedCell(family.Sheet, cell, excludedCells) {
				continue
			}
			memberCells = append(memberCells, cell)
		}
		family.MemberCells = memberCells
		filteredFamilies = append(filteredFamilies, family)
	}
	packet.FormulaFamilies = filteredFamilies

	packet.CitationMap = rebuildCitationMap(packet)
}

func rebuildCitationMap(packet *model.EvidencePacketV1) model.CitationMap {
	sheetNames := map[string]struct{}{}
	sheetCells := map[string]struct{}{}
	ruleIDs := map[string]struct{}{}
	issueIDs := map[string]struct{}{}
	clusterIDs := map[string]struct{}{}

	for _, sheet := range packet.Sheets {
		sheetNames[sheet.Name] = struct{}{}
	}
	for _, issue := range packet.Issues {
		issueIDs[issue.IssueID] = struct{}{}
		ruleIDs[issue.RuleID] = struct{}{}
		if issue.Sheet != "" {
			sheetNames[issue.Sheet] = struct{}{}
		}
		if issue.Sheet != "" && issue.Cell != "" {
			sheetCells[sheetCell(issue.Sheet, issue.Cell)] = struct{}{}
		}
	}
	for _, family := range packet.FormulaFamilies {
		clusterIDs[family.ClusterID] = struct{}{}
		if family.Sheet != "" {
			sheetNames[family.Sheet] = struct{}{}
		}
		for _, cell := range family.MemberCells {
			if family.Sheet != "" && cell != "" {
				sheetCells[sheetCell(family.Sheet, cell)] = struct{}{}
			}
		}
		if family.Sheet != "" && family.OutlierCell != "" {
			sheetCells[sheetCell(family.Sheet, family.OutlierCell)] = struct{}{}
		}
	}

	return model.CitationMap{
		FormulaClusterIDs: sortedKeys(clusterIDs),
		IssueIDs:          sortedKeys(issueIDs),
		RuleIDs:           sortedKeys(ruleIDs),
		SheetCells:        sortedKeys(sheetCells),
		SheetNames:        sortedKeys(sheetNames),
	}
}

func sheetCell(sheet, cell string) string {
	return sheet + "!" + cell
}

func isExcludedCell(sheet, cell string, excludedCells map[string]struct{}) bool {
	_, excluded := excludedCells[sheetCell(sheet, cell)]
	return excluded
}

func filterDetails(sheet string, details map[string]any, excludedCells map[string]struct{}) map[string]any {
	if len(details) == 0 || len(excludedCells) == 0 {
		return details
	}
	filtered := make(map[string]any, len(details))
	for key, value := range details {
		if key == "cluster_cells" {
			value = filterCellList(sheet, value, excludedCells)
		}
		filtered[key] = value
	}
	return filtered
}

func filterCellList(sheet string, value any, excludedCells map[string]struct{}) any {
	switch cells := value.(type) {
	case []string:
		filtered := make([]string, 0, len(cells))
		for _, cell := range cells {
			if isExcludedCell(sheet, cell, excludedCells) {
				continue
			}
			filtered = append(filtered, cell)
		}
		return filtered
	case []any:
		filtered := make([]any, 0, len(cells))
		for _, cell := range cells {
			cellText, ok := cell.(string)
			if ok && isExcludedCell(sheet, cellText, excludedCells) {
				continue
			}
			filtered = append(filtered, cell)
		}
		return filtered
	default:
		return value
	}
}

func stringSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		result[trimmed] = struct{}{}
	}
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
