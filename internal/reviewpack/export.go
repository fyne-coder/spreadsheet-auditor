package reviewpack

import (
	"fmt"
	"os"
	"time"

	"spreadsheet-auditor/internal/model"
)

// Format identifies an export output type.
type Format string

const (
	FormatHTML Format = "html"
	FormatCSV  Format = "csv"
)

// ExportOptions configures deterministic HTML or CSV export from an AuditReport.
type ExportOptions struct {
	Format     Format
	ExportedAt time.Time
	IssueIDs   []string
}

// FilterIssues returns issues whose IssueID is listed when ids is non-empty.
// When ids is empty, all issues are returned. Unknown ids are ignored.
func FilterIssues(issues []model.Issue, ids []string) []model.Issue {
	if len(ids) == 0 {
		return issues
	}
	wanted := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		wanted[id] = struct{}{}
	}
	filtered := make([]model.Issue, 0, len(issues))
	for _, issue := range issues {
		if _, ok := wanted[model.IssueID(issue)]; ok {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

func reportForExport(report *model.AuditReport, opts ExportOptions) *model.AuditReport {
	if len(opts.IssueIDs) == 0 {
		return report
	}
	clone := *report
	clone.Issues = FilterIssues(report.Issues, opts.IssueIDs)
	clone.Summary = clone.ReportSummary()
	return &clone
}

// WriteExport writes HTML or CSV for the report using encoding/csv for CSV output.
func WriteExport(report *model.AuditReport, outputPath string, opts ExportOptions) error {
	switch opts.Format {
	case FormatHTML:
		payload := RenderHTML(reportForExport(report, opts), opts.ExportedAt)
		return os.WriteFile(outputPath, []byte(payload), 0o644)
	case FormatCSV:
		return WriteCSV(reportForExport(report, opts), outputPath, opts.ExportedAt)
	default:
		return fmt.Errorf("unsupported export format %q (use html or csv)", opts.Format)
	}
}

// ParseFormat validates a format string for CLI and service callers.
func ParseFormat(raw string) (Format, error) {
	switch Format(raw) {
	case FormatHTML, FormatCSV:
		return Format(raw), nil
	default:
		return "", fmt.Errorf("unsupported export format %q (use html or csv)", raw)
	}
}
