import { describe, expect, it } from "vitest";
import { model } from "../../wailsjs/go/models";
import { understandingReportToMarkdown } from "./understandingMarkdown";

describe("understandingReportToMarkdown", () => {
  it("renders a verified analysis with section headings and citations", () => {
    const report = new model.UnderstandingReportV1({
      workbook_purpose: [
        new model.UnderstandingClaim({
          claim: "Budget review workbook.",
          citations: ["issue-1"],
        }),
      ],
      sheet_roles: [
        new model.SheetRoleClaim({
          sheet: "Inputs",
          role: "Likely input sheet.",
          citations: ["Inputs"],
        }),
      ],
      key_flows: [],
      major_risks: [
        new model.RiskClaim({
          severity: "medium",
          summary: "External links need owner confirmation.",
          citations: ["EXTERNAL_WORKBOOK_REFERENCE"],
        }),
      ],
      cleanup_plan: [],
      owner_questions: [
        new model.OwnerQuestion({
          question: "Which linked workbook is authoritative?",
          context_citations: ["issue-1"],
        }),
      ],
      confidence_notes: [],
    });

    expect(understandingReportToMarkdown(report)).toContain("# Verified AI Analysis");
    expect(understandingReportToMarkdown(report)).toContain("## Workbook Purpose");
    expect(understandingReportToMarkdown(report)).toContain("- Budget review workbook.");
    expect(understandingReportToMarkdown(report)).toContain("`issue-1`");
    expect(understandingReportToMarkdown(report)).toContain("## Sheet Roles");
    expect(understandingReportToMarkdown(report)).toContain("**Inputs**: Likely input sheet.");
    expect(understandingReportToMarkdown(report)).toContain("## Major Risks");
    expect(understandingReportToMarkdown(report)).toContain("**medium**: External links need owner confirmation.");
    expect(understandingReportToMarkdown(report)).toContain("Context citations: `issue-1`");
  });
});
