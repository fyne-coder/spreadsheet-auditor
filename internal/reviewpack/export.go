package reviewpack

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"spreadsheet-auditor/internal/model"
)

const privateExportFileMode = 0o600

// Format identifies an export output type.
type Format string

const (
	FormatHTML Format = "html"
	FormatCSV  Format = "csv"
)

// ExportOptions configures deterministic HTML or CSV export from an AuditReport.
type ExportOptions struct {
	Format          Format
	ExportedAt      time.Time
	IssueIDs        []string
	IncludeFullPath bool
}

// ExportedWorkbookPath returns the workbook identity string written into exports.
// By default only the basename is exported so local directory names stay private.
func ExportedWorkbookPath(workbookPath string, includeFullPath bool) string {
	if includeFullPath {
		return workbookPath
	}
	return filepath.Base(workbookPath)
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
	exportReport := reportForExport(report, opts)
	workbookIdentity := ExportedWorkbookPath(exportReport.WorkbookPath, opts.IncludeFullPath)
	switch opts.Format {
	case FormatHTML:
		payload := RenderHTML(exportReport, opts.ExportedAt, workbookIdentity)
		return os.WriteFile(outputPath, []byte(payload), privateExportFileMode)
	case FormatCSV:
		return WriteCSV(exportReport, outputPath, opts.ExportedAt, workbookIdentity)
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
