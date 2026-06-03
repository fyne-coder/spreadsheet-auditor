/** Split comma- or newline-separated exclusion tokens from a text field. */
export function parseExclusionList(raw: string): string[] {
  return raw
    .split(/[,\n]+/)
    .map((entry) => entry.trim())
    .filter((entry) => entry.length > 0);
}
