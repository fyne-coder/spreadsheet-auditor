package model

import (
	"bytes"
	"encoding/json"
	"sort"
)

type IssueEvidence struct {
	Sheet   string
	Cell    string
	Formula string
}

type Issue struct {
	RuleID      string
	Title       string
	Severity    string
	Category    string
	Rationale   string
	Remediation string
	Message     string
	Evidence    IssueEvidence
	Details     map[string]any
}

type SheetSummary struct {
	Name         string
	State        string
	UsedRange    string
	FormulaCells int
}

type Summary struct {
	SheetCount       int
	FormulaCellCount int
	IssueCount       int
	IssuesBySeverity map[string]int
	IssuesByCategory map[string]int
}

type AuditReport struct {
	WorkbookPath    string
	SupportedFormat string
	Summary         Summary
	Sheets          []SheetSummary
	Issues          []Issue
}

func BuildIssue(ruleID, message, sheet, cell, formula string, details map[string]any) Issue {
	rule := Rules[ruleID]
	evidence := IssueEvidence{Sheet: sheet, Cell: cell}
	if formula != "" {
		evidence.Formula = formula
	}
	if details == nil {
		details = map[string]any{}
	}
	return Issue{
		RuleID:      rule.RuleID,
		Title:       rule.Title,
		Severity:    rule.Severity,
		Category:    rule.Category,
		Rationale:   rule.Rationale,
		Remediation: rule.Remediation,
		Message:     message,
		Evidence:    evidence,
		Details:     details,
	}
}

// ReportSummary returns aggregate counts derived from sheets and issues.
func (r *AuditReport) ReportSummary() Summary {
	return r.buildSummary()
}

func (r *AuditReport) buildSummary() Summary {
	severityCounts := map[string]int{}
	categoryCounts := map[string]int{}
	formulaCells := 0
	for _, sheet := range r.Sheets {
		formulaCells += sheet.FormulaCells
	}
	for _, issue := range r.Issues {
		severityCounts[issue.Severity]++
		categoryCounts[issue.Category]++
	}
	return Summary{
		SheetCount:       len(r.Sheets),
		FormulaCellCount: formulaCells,
		IssueCount:       len(r.Issues),
		IssuesBySeverity: severityCounts,
		IssuesByCategory: categoryCounts,
	}
}

func IssueSortKey(issue Issue) (string, string, string) {
	return issue.Evidence.Sheet, issue.Evidence.Cell, issue.RuleID
}

func SortIssues(issues []Issue) {
	sort.Slice(issues, func(i, j int) bool {
		si, ci, ri := IssueSortKey(issues[i])
		sj, cj, rj := IssueSortKey(issues[j])
		if si != sj {
			return si < sj
		}
		if ci != cj {
			return ci < cj
		}
		return ri < rj
	})
}

func sortedStringKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedAnyKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type evidenceJSON struct {
	Cell    string  `json:"cell"`
	Formula *string `json:"formula,omitempty"`
	Sheet   string  `json:"sheet"`
}

type issueJSON struct {
	Category    string         `json:"category"`
	Details     map[string]any `json:"details,omitempty"`
	Evidence    evidenceJSON   `json:"evidence"`
	Message     string         `json:"message"`
	Rationale   string         `json:"rationale"`
	Remediation string         `json:"remediation"`
	RuleID      string         `json:"rule_id"`
	Severity    string         `json:"severity"`
	Title       string         `json:"title"`
}

type sheetJSON struct {
	FormulaCells int    `json:"formula_cells"`
	Name         string `json:"name"`
	State        string `json:"state"`
	UsedRange    string `json:"used_range"`
}

type summaryJSON struct {
	FormulaCellCount int            `json:"formula_cell_count"`
	IssueCount       int            `json:"issue_count"`
	IssuesByCategory map[string]int `json:"issues_by_category"`
	IssuesBySeverity map[string]int `json:"issues_by_severity"`
	SheetCount       int            `json:"sheet_count"`
}

type reportJSON struct {
	Issues          []issueJSON `json:"issues"`
	Sheets          []sheetJSON `json:"sheets"`
	Summary         summaryJSON `json:"summary"`
	SupportedFormat string      `json:"supported_format"`
	WorkbookPath    string      `json:"workbook_path"`
}

func issueToJSON(issue Issue) issueJSON {
	evidence := evidenceJSON{
		Cell:  issue.Evidence.Cell,
		Sheet: issue.Evidence.Sheet,
	}
	if issue.Evidence.Formula != "" {
		formula := issue.Evidence.Formula
		evidence.Formula = &formula
	}
	payload := issueJSON{
		Category:    issue.Category,
		Evidence:    evidence,
		Message:     issue.Message,
		Rationale:   issue.Rationale,
		Remediation: issue.Remediation,
		RuleID:      issue.RuleID,
		Severity:    issue.Severity,
		Title:       issue.Title,
	}
	if len(issue.Details) > 0 {
		details := make(map[string]any, len(issue.Details))
		for _, key := range sortedAnyKeys(issue.Details) {
			details[key] = issue.Details[key]
		}
		payload.Details = details
	}
	return payload
}

func sortedCountMap(source map[string]int) map[string]int {
	if len(source) == 0 {
		return map[string]int{}
	}
	result := make(map[string]int, len(source))
	for _, key := range sortedStringKeys(source) {
		result[key] = source[key]
	}
	return result
}

func (r *AuditReport) reportPayload() reportJSON {
	summary := r.buildSummary()
	issues := make([]issueJSON, 0, len(r.Issues))
	for _, issue := range r.Issues {
		issues = append(issues, issueToJSON(issue))
	}
	sheets := make([]sheetJSON, 0, len(r.Sheets))
	for _, sheet := range r.Sheets {
		sheets = append(sheets, sheetJSON{
			FormulaCells: sheet.FormulaCells,
			Name:         sheet.Name,
			State:        sheet.State,
			UsedRange:    sheet.UsedRange,
		})
	}
	return reportJSON{
		Issues: issues,
		Sheets: sheets,
		Summary: summaryJSON{
			FormulaCellCount: summary.FormulaCellCount,
			IssueCount:       summary.IssueCount,
			IssuesByCategory: sortedCountMap(summary.IssuesByCategory),
			IssuesBySeverity: sortedCountMap(summary.IssuesBySeverity),
			SheetCount:       summary.SheetCount,
		},
		SupportedFormat: r.SupportedFormat,
		WorkbookPath:    r.WorkbookPath,
	}
}

// CanonicalJSON matches Python json.dumps(..., indent=2, sort_keys=True) + trailing newline.
func (r *AuditReport) CanonicalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(r.reportPayload()); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
