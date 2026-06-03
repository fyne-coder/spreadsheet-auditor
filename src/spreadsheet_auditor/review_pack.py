from __future__ import annotations

from html import escape

from spreadsheet_auditor.models import AuditReport, Issue, SheetSummary


def render_review_pack_html(report: AuditReport) -> str:
    """Render a manager-readable HTML review pack from an AuditReport."""
    summary = report.to_dict()["summary"]
    severity_rows = _count_rows(summary["issues_by_severity"])
    category_rows = _count_rows(summary["issues_by_category"])
    sheet_rows = _sheet_rows(report.sheets)
    issue_rows = _issue_rows(report.issues)

    workbook_path = escape(str(report.workbook_path))
    supported_format = escape(report.supported_format)

    return f"""<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Spreadsheet Auditor Review Pack</title>
  <style>
    body {{
      font-family: system-ui, -apple-system, sans-serif;
      line-height: 1.5;
      margin: 2rem auto;
      max-width: 1100px;
      padding: 0 1rem;
      color: #1a1a1a;
    }}
    h1, h2 {{ margin-top: 2rem; }}
    table {{
      border-collapse: collapse;
      margin: 1rem 0;
      width: 100%;
    }}
    th, td {{
      border: 1px solid #ccc;
      padding: 0.5rem 0.75rem;
      text-align: left;
      vertical-align: top;
    }}
    th {{ background: #f4f4f4; }}
    .meta {{ color: #444; }}
    .formula {{
      font-family: ui-monospace, monospace;
      white-space: pre-wrap;
      word-break: break-word;
    }}
    .severity-high {{ color: #9b1c1c; }}
    .severity-medium {{ color: #92400e; }}
    .severity-low {{ color: #1e40af; }}
  </style>
</head>
<body>
  <h1>Spreadsheet Auditor Review Pack</h1>
  <p class="meta"><strong>Workbook:</strong> {workbook_path}</p>
  <p class="meta"><strong>Format:</strong> {supported_format}</p>

  <h2>Workbook Summary</h2>
  <table>
    <tbody>
      <tr><th scope="row">Sheets</th><td>{summary["sheet_count"]}</td></tr>
      <tr><th scope="row">Formula cells</th><td>{summary["formula_cell_count"]}</td></tr>
      <tr><th scope="row">Issues</th><td>{summary["issue_count"]}</td></tr>
    </tbody>
  </table>

  <h2>Issues by Severity</h2>
  {_table(["Severity", "Count"], severity_rows)}

  <h2>Issues by Category</h2>
  {_table(["Category", "Count"], category_rows)}

  <h2>Sheet Inventory</h2>
  {_table(["Sheet", "State", "Used range", "Formula cells"], sheet_rows)}

  <h2>Issues</h2>
  {_table(
      [
          "Severity",
          "Category",
          "Rule",
          "Sheet",
          "Cell",
          "Formula",
          "Message",
          "Remediation",
      ],
      issue_rows,
  )}
</body>
</html>
"""


def _count_rows(counts: dict[str, int]) -> list[list[str]]:
    if not counts:
        return [["(none)", "0"]]
    return [[escape(key), str(counts[key])] for key in sorted(counts)]


def _sheet_rows(sheets: list[SheetSummary]) -> list[list[str]]:
    if not sheets:
        return [["(none)", "", "", "0"]]
    return [
        [
            escape(sheet.name),
            escape(sheet.state),
            escape(sheet.used_range),
            str(sheet.formula_cells),
        ]
        for sheet in sheets
    ]


def _issue_rows(issues: list[Issue]) -> list[list[str]]:
    if not issues:
        return [["(none)", "", "", "", "", "", "", ""]]
    rows: list[list[str]] = []
    for issue in issues:
        formula = issue.evidence.formula or ""
        rows.append(
            [
                _severity_cell(issue.severity),
                escape(issue.category),
                escape(f"{issue.rule_id}: {issue.title}"),
                escape(issue.evidence.sheet),
                escape(issue.evidence.cell),
                f'<span class="formula">{escape(formula)}</span>' if formula else "",
                escape(issue.message),
                escape(issue.remediation),
            ]
        )
    return rows


def _severity_cell(severity: str) -> str:
    css_class = f"severity-{severity}" if severity in {"high", "medium", "low"} else ""
    label = escape(severity)
    if css_class:
        return f'<span class="{css_class}">{label}</span>'
    return label


def _table(headers: list[str], rows: list[list[str]]) -> str:
    header_html = "".join(f"<th scope=\"col\">{escape(header)}</th>" for header in headers)
    body_html = "".join(
        "<tr>" + "".join(f"<td>{cell}</td>" for cell in row) + "</tr>" for row in rows
    )
    return f"<table><thead><tr>{header_html}</tr></thead><tbody>{body_html}</tbody></table>"
