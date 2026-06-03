package reviewpack

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"spreadsheet-auditor/internal/model"
)

// RenderHTML renders a manager-readable HTML review pack from an AuditReport.
// exportedAt is shown in RFC3339 UTC when non-zero.
func RenderHTML(report *model.AuditReport, exportedAt time.Time) string {
	summary := report.ReportSummary()
	severityRows := countRows(summary.IssuesBySeverity)
	categoryRows := countRows(summary.IssuesByCategory)
	sheetRows := sheetRows(report.Sheets)
	issueRows := issueRows(report.Issues)

	workbookPath := html.EscapeString(report.WorkbookPath)
	supportedFormat := html.EscapeString(report.SupportedFormat)

	var builder strings.Builder
	builder.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Spreadsheet Auditor Review Pack</title>
  <style>
    body {
      font-family: system-ui, -apple-system, sans-serif;
      line-height: 1.5;
      margin: 2rem auto;
      max-width: 1100px;
      padding: 0 1rem;
      color: #1a1a1a;
    }
    h1, h2 { margin-top: 2rem; }
    table {
      border-collapse: collapse;
      margin: 1rem 0;
      width: 100%;
    }
    th, td {
      border: 1px solid #ccc;
      padding: 0.5rem 0.75rem;
      text-align: left;
      vertical-align: top;
    }
    th { background: #f4f4f4; }
    .meta { color: #444; }
    .formula {
      font-family: ui-monospace, monospace;
      white-space: pre-wrap;
      word-break: break-word;
    }
    .severity-high { color: #9b1c1c; }
    .severity-medium { color: #92400e; }
    .severity-low { color: #1e40af; }
  </style>
</head>
<body>
  <h1>Spreadsheet Auditor Review Pack</h1>
  <p class="meta"><strong>Workbook:</strong> `)
	builder.WriteString(workbookPath)
	builder.WriteString(`</p>
  <p class="meta"><strong>Format:</strong> `)
	builder.WriteString(supportedFormat)
	builder.WriteString(`</p>`)
	if !exportedAt.IsZero() {
		builder.WriteString(`
  <p class="meta"><strong>Exported at:</strong> `)
		builder.WriteString(html.EscapeString(exportedAt.UTC().Format(time.RFC3339)))
		builder.WriteString(`</p>`)
	}
	builder.WriteString(`

  <h2>Workbook Summary</h2>
  <table>
    <tbody>
      <tr><th scope="row">Sheets</th><td>`)
	fmt.Fprintf(&builder, "%d", summary.SheetCount)
	builder.WriteString(`</td></tr>
      <tr><th scope="row">Formula cells</th><td>`)
	fmt.Fprintf(&builder, "%d", summary.FormulaCellCount)
	builder.WriteString(`</td></tr>
      <tr><th scope="row">Issues</th><td>`)
	fmt.Fprintf(&builder, "%d", summary.IssueCount)
	builder.WriteString(`</td></tr>
    </tbody>
  </table>

  <h2>Issues by Severity</h2>
  `)
	builder.WriteString(table([]string{"Severity", "Count"}, severityRows))
	builder.WriteString(`

  <h2>Issues by Category</h2>
  `)
	builder.WriteString(table([]string{"Category", "Count"}, categoryRows))
	builder.WriteString(`

  <h2>Sheet Inventory</h2>
  `)
	builder.WriteString(table([]string{"Sheet", "State", "Used range", "Formula cells"}, sheetRows))
	builder.WriteString(`

  <h2>Issues</h2>
  `)
	builder.WriteString(table(
		[]string{
			"Severity",
			"Category",
			"Rule",
			"Sheet",
			"Cell",
			"Formula",
			"Message",
			"Remediation",
		},
		issueRows,
	))
	builder.WriteString(`
</body>
</html>
`)
	return builder.String()
}

func countRows(counts map[string]int) [][]string {
	if len(counts) == 0 {
		return [][]string{{"(none)", "0"}}
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := make([][]string, 0, len(keys))
	for _, key := range keys {
		rows = append(rows, []string{html.EscapeString(key), fmt.Sprintf("%d", counts[key])})
	}
	return rows
}

func sheetRows(sheets []model.SheetSummary) [][]string {
	if len(sheets) == 0 {
		return [][]string{{"(none)", "", "", "0"}}
	}
	rows := make([][]string, 0, len(sheets))
	for _, sheet := range sheets {
		rows = append(rows, []string{
			html.EscapeString(sheet.Name),
			html.EscapeString(sheet.State),
			html.EscapeString(sheet.UsedRange),
			fmt.Sprintf("%d", sheet.FormulaCells),
		})
	}
	return rows
}

func issueRows(issues []model.Issue) [][]string {
	if len(issues) == 0 {
		return [][]string{{"(none)", "", "", "", "", "", "", ""}}
	}
	rows := make([][]string, 0, len(issues))
	for _, issue := range issues {
		formula := issue.Evidence.Formula
		formulaCell := ""
		if formula != "" {
			formulaCell = fmt.Sprintf(
				`<span class="formula">%s</span>`,
				html.EscapeString(formula),
			)
		}
		rows = append(rows, []string{
			severityCell(issue.Severity),
			html.EscapeString(issue.Category),
			html.EscapeString(fmt.Sprintf("%s: %s", issue.RuleID, issue.Title)),
			html.EscapeString(issue.Evidence.Sheet),
			html.EscapeString(issue.Evidence.Cell),
			formulaCell,
			html.EscapeString(issue.Message),
			html.EscapeString(issue.Remediation),
		})
	}
	return rows
}

func severityCell(severity string) string {
	switch severity {
	case "high", "medium", "low":
		return fmt.Sprintf(
			`<span class="severity-%s">%s</span>`,
			severity,
			html.EscapeString(severity),
		)
	default:
		return html.EscapeString(severity)
	}
}

func table(headers []string, rows [][]string) string {
	var builder strings.Builder
	builder.WriteString("<table><thead><tr>")
	for _, header := range headers {
		builder.WriteString(`<th scope="col">`)
		builder.WriteString(html.EscapeString(header))
		builder.WriteString("</th>")
	}
	builder.WriteString("</tr></thead><tbody>")
	for _, row := range rows {
		builder.WriteString("<tr>")
		for _, cell := range row {
			builder.WriteString("<td>")
			builder.WriteString(cell)
			builder.WriteString("</td>")
		}
		builder.WriteString("</tr>")
	}
	builder.WriteString("</tbody></table>")
	return builder.String()
}
