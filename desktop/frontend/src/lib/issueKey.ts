import type { model } from "../../wailsjs/go/models";

/** Backend issue key for analyst handoff, with a fallback for hand-built fixtures. */
export function issueKey(issue: model.Issue): string {
  if (issue.IssueID) {
    return issue.IssueID;
  }
  const sheet = issue.Evidence?.Sheet ?? "";
  const cell = issue.Evidence?.Cell ?? "";
  return `${issue.RuleID}|${sheet}|${cell}|${issue.Message}`;
}
