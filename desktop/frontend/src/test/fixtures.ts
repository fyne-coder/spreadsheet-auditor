import { model } from "../../wailsjs/go/models";

export function makeIssue(
  overrides: Partial<model.Issue> & Pick<model.Issue, "RuleID" | "Message">,
): model.Issue {
  return new model.Issue({
    RuleID: overrides.RuleID,
    Title: overrides.Title ?? overrides.RuleID,
    Severity: overrides.Severity ?? "medium",
    Category: overrides.Category ?? "formula",
    Rationale: overrides.Rationale ?? "Test rationale",
    Remediation: overrides.Remediation ?? "Fix it",
    Message: overrides.Message,
    Evidence: overrides.Evidence ?? new model.IssueEvidence({ Sheet: "Sheet1", Cell: "A1" }),
    Details: overrides.Details ?? {},
  });
}

export function makeReport(issues: model.Issue[]): model.AuditReport {
  const bySeverity: Record<string, number> = {};
  const byCategory: Record<string, number> = {};
  for (const issue of issues) {
    bySeverity[issue.Severity] = (bySeverity[issue.Severity] ?? 0) + 1;
    byCategory[issue.Category] = (byCategory[issue.Category] ?? 0) + 1;
  }
  return new model.AuditReport({
    WorkbookPath: "/tmp/sample.xlsx",
    SupportedFormat: "xlsx",
    Summary: new model.Summary({
      SheetCount: 1,
      FormulaCellCount: 10,
      IssueCount: issues.length,
      IssuesBySeverity: bySeverity,
      IssuesByCategory: byCategory,
    }),
    Sheets: [],
    Issues: issues,
  });
}
