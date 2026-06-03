import { Button, Group, Paper, SimpleGrid, Stack, Text, Title } from "@mantine/core";
import type { ColumnFiltersState } from "@tanstack/react-table";
import type { model } from "../../../wailsjs/go/models";
import { priorityBuckets } from "./priorityBuckets";

type Props = {
  report: model.AuditReport;
  onApplyFilters: (filters: ColumnFiltersState) => void;
  onStartAIHandoff: () => void;
};

function readableCategory(category: string) {
  return category.replaceAll("_", " ");
}

function topCategories(report: model.AuditReport) {
  return Object.entries(report.Summary.IssuesByCategory ?? {})
    .sort(([, a], [, b]) => b - a)
    .slice(0, 3)
    .map(([category, count]) => `${count} ${readableCategory(category)}`);
}

export function OverviewPanel({ report, onApplyFilters, onStartAIHandoff }: Props) {
  const buckets = priorityBuckets(report.Issues ?? []);
  const categorySummary = topCategories(report);

  return (
    <Paper withBorder p="sm" radius="md" data-testid="overview-panel">
      <Stack gap="sm">
        <Group justify="space-between" align="flex-start" wrap="wrap">
          <Stack gap={2} style={{ flex: "1 1 560px" }}>
            <Title order={4}>Audit overview</Title>
            <Text size="sm" c="dimmed">
              The app checks workbook structure and stored formulas. It does not run
              macros, refresh links, recalculate formulas, or decide whether the
              workbook's numbers are right.
            </Text>
          </Stack>
          <Button
            variant="light"
            size="compact-sm"
            onClick={onStartAIHandoff}
            data-testid="overview-ai-handoff"
            aria-label="Open AI review package"
            style={{ flexShrink: 0 }}
          >
            AI package
          </Button>
        </Group>

        <Text size="sm">
          Found <strong>{report.Summary.IssueCount}</strong> issues across{" "}
          <strong>{report.Summary.SheetCount}</strong> sheets
          {categorySummary.length > 0 ? `, including ${categorySummary.join(", ")}.` : "."}
        </Text>

        <SimpleGrid cols={{ base: 1, sm: 2, md: 4 }} spacing="xs">
          {buckets.map((bucket) => (
            <Paper key={bucket.id} withBorder p="xs" radius="sm">
              <Stack gap={6}>
                <Text fw={700} size="sm">
                  {bucket.title}
                </Text>
                <Text size="xl" fw={800}>
                  {bucket.count}
                </Text>
                <Text size="xs" c="dimmed" mih={44}>
                  {bucket.description}
                </Text>
                <Button
                  size="compact-xs"
                  variant={bucket.count > 0 ? "light" : "default"}
                  disabled={bucket.count === 0}
                  onClick={() => onApplyFilters(bucket.filters)}
                  data-testid={`overview-bucket-${bucket.id}`}
                >
                  Show in issues
                </Button>
              </Stack>
            </Paper>
          ))}
        </SimpleGrid>

        <Text size="xs" c="dimmed">
          These paths are meant to route review effort. A finding can appear in more
          than one path when the same issue affects severity and workbook lineage.
        </Text>

        <Paper p="xs" radius="sm" bg="gray.0">
          <Text fw={700} size="sm">
            What the app cannot tell you
          </Text>
          <Text size="sm" c="dimmed">
            The scanner can flag risky formulas and workbook structure, but it cannot
            infer full business purpose. For a fuller explanation, copy the optional
            AI review package and paste it into the assistant you already use.
          </Text>
        </Paper>
      </Stack>
    </Paper>
  );
}
