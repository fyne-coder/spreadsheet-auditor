import { describe, expect, it } from "vitest";
import { makeIssue } from "../test/fixtures";
import { issueKey } from "./issueKey";

describe("issueKey", () => {
  it("uses the backend issue id when present", () => {
    const issue = makeIssue({
      IssueID: "RULE_A|S1|A1|same message|abc12345",
      RuleID: "RULE_A",
      Message: "same message",
      Evidence: { Sheet: "S1", Cell: "A1", Formula: "" },
    });
    expect(issueKey(issue)).toBe("RULE_A|S1|A1|same message|abc12345");
  });

  it("falls back deterministically for hand-built fixtures", () => {
    const issue = makeIssue({
      RuleID: "RULE_A",
      Message: "same message",
      Evidence: { Sheet: "S1", Cell: "A1", Formula: "" },
    });
    expect(issueKey(issue)).toBe("RULE_A|S1|A1|same message");
    expect(issueKey(issue)).toBe(issueKey(issue));
  });
});
