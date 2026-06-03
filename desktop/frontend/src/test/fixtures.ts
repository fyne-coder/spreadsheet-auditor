import { model } from "../../wailsjs/go/models";

export function makeIssue(
  overrides: Partial<model.Issue> & Pick<model.Issue, "RuleID" | "Message">,
): model.Issue {
  return new model.Issue({
    IssueID: overrides.IssueID ?? "",
    RuleID: overrides.RuleID,
    Title: overrides.Title ?? overrides.RuleID,
    Severity: overrides.Severity ?? "medium",
    Category: overrides.Category ?? "formula",
    Priority: overrides.Priority ?? "",
    ImpactFactors: overrides.ImpactFactors ?? [],
    Rationale: overrides.Rationale ?? "Test rationale",
    Remediation: overrides.Remediation ?? "Fix it",
    Message: overrides.Message,
    Evidence: overrides.Evidence ?? new model.IssueEvidence({ Sheet: "Sheet1", Cell: "A1" }),
    Details: overrides.Details ?? {},
  });
}

export function makePromptBundle(overrides?: {
  prompt?: string;
  issueId?: string;
}): model.PromptBundleV1 {
  const issueId = overrides?.issueId ?? "issue-1";
  return new model.PromptBundleV1({
    bundle_version: "1",
    prompt_version: "1",
    instructions: "test instructions",
    response_schema: {},
    prompt: overrides?.prompt ?? "prompt with exclusions",
    evidence_packet: new model.EvidencePacketV1({
      audit_hash: "abc",
      packet_version: "1",
      workbook: new model.EvidenceWorkbook({
        name: "sample.xlsx",
        supported_format: "xlsx",
        summary: new model.Summary({
          SheetCount: 1,
          FormulaCellCount: 1,
          IssueCount: 1,
          IssuesBySeverity: { medium: 1 },
          IssuesByCategory: { formula: 1 },
        }),
      }),
      sheets: [new model.EvidenceSheet({ name: "Model", formula_cells: 1, state: "visible", used_range: "A1" })],
      audit_findings: [
        new model.EvidenceIssue({
          issue_id: issueId,
          rule_id: "RULE_A",
          title: "Issue",
          severity: "medium",
          category: "formula",
          sheet: "Model",
          cell: "B1",
          message: "msg",
          rationale: "why",
          remediation: "fix",
        }),
      ],
      formula_families: [],
      citation_map: new model.CitationMap({
        issue_ids: [issueId],
        rule_ids: [],
        sheet_cells: ["Model!B1"],
        sheet_names: ["Model"],
        formula_cluster_ids: [],
      }),
    }),
  });
}

export function makeUnderstandingReport(issueId: string): model.UnderstandingReportV1 {
  return new model.UnderstandingReportV1({
    workbook_purpose: [new model.UnderstandingClaim({ claim: "Purpose", citations: [issueId] })],
    sheet_roles: [],
    key_flows: [],
    major_risks: [],
    cleanup_plan: [],
    owner_questions: [],
    confidence_notes: [],
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
