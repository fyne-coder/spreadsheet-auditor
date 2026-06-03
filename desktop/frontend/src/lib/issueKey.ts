import type { model } from "../../wailsjs/go/models";

/** Deterministic issue key for analyst handoff (no backend issue_id yet). */
export function issueKey(issue: model.Issue): string {
  const sheet = issue.Evidence?.Sheet ?? "";
  const cell = issue.Evidence?.Cell ?? "";
  return `${issue.RuleID}|${sheet}|${cell}|${issue.Message}`;
}
