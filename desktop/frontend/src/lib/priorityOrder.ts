const priorityRank: Record<string, number> = {
  critical: 4,
  high: 3,
  medium: 2,
  low: 1,
};

export function comparePriority(left: string, right: string) {
  return (priorityRank[right] ?? 0) - (priorityRank[left] ?? 0);
}

export function impactFactorCodes(
  factors: Array<{ Code?: string; code?: string }> | undefined,
) {
  if (!factors?.length) {
    return "";
  }
  return factors
    .map((factor) => factor.Code ?? factor.code ?? "")
    .filter(Boolean)
    .join(", ");
}
