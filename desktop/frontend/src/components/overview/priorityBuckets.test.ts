import { describe, expect, it } from "vitest";
import { makeIssue } from "../../test/fixtures";
import { priorityBuckets } from "./priorityBuckets";

describe("priorityBuckets", () => {
  it("groups issues by derived analyst priority bands", () => {
    const buckets = priorityBuckets([
      makeIssue({
        RuleID: "BROKEN_REF_FORMULA",
        Severity: "high",
        Priority: "critical",
        Category: "formula_integrity",
        Message: "broken",
      }),
      makeIssue({
        RuleID: "EXTERNAL_WORKBOOK_REFERENCE",
        Severity: "medium",
        Priority: "high",
        Category: "lineage",
        Message: "external",
      }),
      makeIssue({
        RuleID: "VOLATILE_FUNCTION",
        Severity: "medium",
        Priority: "medium",
        Category: "performance",
        Message: "volatile",
      }),
      makeIssue({
        RuleID: "VOLATILE_FUNCTION",
        Severity: "medium",
        Priority: "low",
        Category: "performance",
        Message: "today only",
      }),
    ]);

    expect(buckets.map((bucket) => [bucket.id, bucket.count])).toEqual([
      ["critical", 1],
      ["high", 1],
      ["medium", 1],
      ["low", 1],
    ]);
    expect(buckets[0].filters).toEqual([{ id: "Priority", value: ["critical"] }]);
    expect(buckets[3].filters).toEqual([{ id: "Priority", value: ["low"] }]);
  });
});
