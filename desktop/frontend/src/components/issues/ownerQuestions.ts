import type { model } from "../../../wailsjs/go/models";

const ownerQuestionByRule: Record<string, string> = {
  BROKEN_REF_FORMULA: "What range or sheet should this broken reference point to?",
  EXTERNAL_WORKBOOK_REFERENCE: "Who owns the linked workbook, and should this file depend on it?",
  HARDCODED_NUMERIC_CONSTANT: "Is this number an approved assumption that should live in a labeled input cell?",
  VOLATILE_FUNCTION: "Does this value need to recalculate every time the workbook changes?",
  WHOLE_COLUMN_RANGE: "Can this formula use a bounded range or table instead of a full column?",
};

export function ownerQuestion(issue: model.Issue): string {
  return (
    ownerQuestionByRule[issue.RuleID] ??
    "Ask the previous owner whether this issue is intentional and who should approve the fix."
  );
}
