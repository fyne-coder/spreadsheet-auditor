import type { ColumnFiltersState } from "@tanstack/react-table";
import type { model } from "../../../wailsjs/go/models";

export type PriorityBucket = {
  id: "critical" | "high" | "medium" | "low";
  title: string;
  description: string;
  count: number;
  filters: ColumnFiltersState;
};

const bucketOrder: PriorityBucket["id"][] = ["critical", "high", "medium", "low"];

const bucketMeta: Record<
  PriorityBucket["id"],
  Pick<PriorityBucket, "title" | "description">
> = {
  critical: {
    title: "Critical routing",
    description: "Hidden-sheet structural defects and other issues that need immediate review.",
  },
  high: {
    title: "High priority",
    description: "Broken references, external links, and higher-risk volatile formulas.",
  },
  medium: {
    title: "Review soon",
    description: "Formula hygiene, moderate clusters, and other medium routing work.",
  },
  low: {
    title: "Cleanup later",
    description: "Isolated low-risk volatility and other lower-priority follow-up.",
  },
};

function priorityCount(issues: model.Issue[], band: PriorityBucket["id"]) {
  return issues.filter((issue) => issue.Priority === band).length;
}

export function priorityBuckets(issues: model.Issue[]): PriorityBucket[] {
  return bucketOrder.map((id) => ({
    id,
    ...bucketMeta[id],
    count: priorityCount(issues, id),
    filters: [{ id: "Priority", value: [id] }],
  }));
}
