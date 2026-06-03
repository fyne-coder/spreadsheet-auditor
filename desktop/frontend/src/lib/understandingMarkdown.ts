import type { model } from "../../wailsjs/go/models";

function cleanText(value: string | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim();
}

function citationText(citations: string[] | undefined, label = "Citations"): string {
  const resolved = (citations ?? []).map(cleanText).filter(Boolean);
  if (resolved.length === 0) {
    return `  - ${label}: none`;
  }
  return `  - ${label}: ${resolved.map((citation) => `\`${citation}\``).join(", ")}`;
}

function pushSection(lines: string[], title: string, items: string[]) {
  if (items.length === 0) {
    return;
  }
  lines.push("", `## ${title}`, "", ...items);
}

export function understandingReportToMarkdown(report: model.UnderstandingReportV1): string {
  const lines = [
    "# Verified AI Analysis",
    "",
    "Cited evidence check passed: every cited ID resolved to the workbook evidence packet.",
    "Review the AI's interpretation before relying on it; this is not a recalculated workbook result.",
  ];

  pushSection(
    lines,
    "Workbook Purpose",
    (report.workbook_purpose ?? []).map(
      (item) => `- ${cleanText(item.claim)}\n${citationText(item.citations)}`,
    ),
  );
  pushSection(
    lines,
    "Sheet Roles",
    (report.sheet_roles ?? []).map(
      (item) =>
        `- **${cleanText(item.sheet)}**: ${cleanText(item.role)}\n${citationText(item.citations)}`,
    ),
  );
  pushSection(
    lines,
    "Key Flows",
    (report.key_flows ?? []).map(
      (item) => `- ${cleanText(item.summary)}\n${citationText(item.citations)}`,
    ),
  );
  pushSection(
    lines,
    "Major Risks",
    (report.major_risks ?? []).map(
      (item) =>
        `- **${cleanText(item.severity)}**: ${cleanText(item.summary)}\n${citationText(item.citations)}`,
    ),
  );
  pushSection(
    lines,
    "Cleanup Plan",
    (report.cleanup_plan ?? []).map(
      (item) => `- ${cleanText(item.action)}\n${citationText(item.citations)}`,
    ),
  );
  pushSection(
    lines,
    "Owner Questions",
    (report.owner_questions ?? []).map(
      (item) =>
        `- ${cleanText(item.question)}\n${citationText(item.context_citations, "Context citations")}`,
    ),
  );
  pushSection(
    lines,
    "Confidence Notes",
    (report.confidence_notes ?? []).map(
      (item) => `- ${cleanText(item.note)}\n${citationText(item.citations)}`,
    ),
  );

  if (lines.length === 4) {
    lines.push("", "No cited sections returned.");
  }

  return `${lines.join("\n")}\n`;
}
