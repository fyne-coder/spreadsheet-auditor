export type ExportFormat = "html" | "csv";

export type ExportScope = "all" | "selected";

export function defaultExportFilename(format: ExportFormat): string {
  return format === "csv" ? "review-pack.csv" : "review-pack.html";
}

export function formatSegmentLabel(format: ExportFormat): string {
  return format === "csv" ? "CSV" : "HTML";
}
