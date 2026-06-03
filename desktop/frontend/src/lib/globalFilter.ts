import type { model } from "../../wailsjs/go/models";

export function issueSearchText(issue: model.Issue): string {
  const parts = [
    issue.RuleID,
    issue.Title,
    issue.Message,
    issue.Rationale,
    issue.Remediation,
    issue.Severity,
    issue.Category,
    issue.Evidence?.Sheet,
    issue.Evidence?.Cell,
    issue.Evidence?.Formula,
  ];
  return parts.filter(Boolean).join(" ").toLowerCase();
}

export function matchesGlobalFilter(
  issue: model.Issue,
  filter: string,
): boolean {
  const query = filter.trim().toLowerCase();
  if (!query) {
    return true;
  }
  return issueSearchText(issue).includes(query);
}
