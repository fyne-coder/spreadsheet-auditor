package reviewpack

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"spreadsheet-auditor/internal/model"
)

var csvHeader = []string{
	"exported_at",
	"workbook_path",
	"supported_format",
	"issue_id",
	"priority",
	"impact_factors",
	"severity",
	"category",
	"rule_id",
	"title",
	"sheet",
	"cell",
	"formula",
	"message",
	"rationale",
	"remediation",
	"details_json",
}

// WriteCSV writes a review-pack CSV using encoding/csv.
func WriteCSV(report *model.AuditReport, outputPath string, exportedAt time.Time, workbookIdentity string) error {
	if workbookIdentity == "" {
		workbookIdentity = ExportedWorkbookPath(report.WorkbookPath, false)
	}
	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, privateExportFileMode)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(csvHeader); err != nil {
		return err
	}

	exportedAtValue := exportedAt.UTC().Format(time.RFC3339)
	for _, issue := range report.Issues {
		detailsJSON, err := CanonicalDetailsJSON(issue.Details)
		if err != nil {
			return fmt.Errorf("encode details_json for %s: %w", model.IssueID(issue), err)
		}
		row := []string{
			exportedAtValue,
			workbookIdentity,
			report.SupportedFormat,
			model.IssueID(issue),
			issue.Priority,
			impactFactorSummary(issue.ImpactFactors),
			issue.Severity,
			issue.Category,
			issue.RuleID,
			issue.Title,
			issue.Evidence.Sheet,
			issue.Evidence.Cell,
			issue.Evidence.Formula,
			issue.Message,
			issue.Rationale,
			issue.Remediation,
			detailsJSON,
		}
		for index, value := range row {
			row[index] = sanitizeCSVField(value)
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// sanitizeCSVField prefixes Excel-formula-like values per the CSV injection policy.
func sanitizeCSVField(value string) string {
	if value == "" {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	default:
		return value
	}
}

const csvHeaderLine = "exported_at,workbook_path,supported_format,issue_id,priority,impact_factors,severity,category,rule_id,title,sheet,cell,formula,message,rationale,remediation,details_json\n"

// CSVHeaderLine returns the canonical header row bytes for regression tests.
func CSVHeaderLine() string {
	return csvHeaderLine
}
