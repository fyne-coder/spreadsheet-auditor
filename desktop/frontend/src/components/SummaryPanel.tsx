import { Badge, Button, Group, Paper, SimpleGrid, Stack, Text } from "@mantine/core";
import { useState } from "react";
import type { model } from "../../wailsjs/go/models";
import { splitPath } from "../lib/splitPath";

type Props = {
  report: model.AuditReport;
};

function RollupChips({ counts }: { counts: Record<string, number> | undefined }) {
  const entries = Object.entries(counts ?? {}).sort(([a], [b]) => a.localeCompare(b));
  if (entries.length === 0) {
    return (
      <Badge variant="light" color="gray" size="xs">
        None
      </Badge>
    );
  }
  return (
    <Group gap={6}>
      {entries.map(([key, value]) => (
        <Badge key={key} variant="light" size="xs">
          {key}: {value}
        </Badge>
      ))}
    </Group>
  );
}

export function SummaryPanel({ report }: Props) {
  const [breakdownOpen, setBreakdownOpen] = useState(false);
  const { filename, directory } = splitPath(report.WorkbookPath ?? "");
  const summary = report.Summary;
  const issuesBySheet = (report.Issues ?? []).reduce<Record<string, number>>((counts, issue) => {
    const sheet = issue.Evidence?.Sheet || "(unknown)";
    counts[sheet] = (counts[sheet] ?? 0) + 1;
    return counts;
  }, {});

  return (
    <Paper withBorder p="xs" radius="md" data-testid="summary-panel">
      <Group justify="space-between" align="flex-start" mb={4} wrap="nowrap">
        <Stack gap={2} style={{ minWidth: 0 }}>
          <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
            Workbook
          </Text>
          <Text fw={700} size="md" lineClamp={1}>
            {filename || "Unknown workbook"}
          </Text>
          {directory ? (
            <Text size="xs" c="dimmed" lineClamp={1}>
              {directory}
            </Text>
          ) : null}
        </Stack>
        <Badge variant="outline" size="sm">
          {report.SupportedFormat}
        </Badge>
      </Group>

      <Group justify="space-between" align="center" gap="xs" wrap="nowrap">
        <Group gap="lg" wrap="wrap">
          <Text size="xs" c="dimmed">
            Issues <Text span fw={700} c="dark">{summary.IssueCount}</Text>
          </Text>
          <Text size="xs" c="dimmed">
            High <Text span fw={700} c="dark">{summary.IssuesBySeverity?.high ?? 0}</Text>
          </Text>
          <Text size="xs" c="dimmed">
            Medium <Text span fw={700} c="dark">{summary.IssuesBySeverity?.medium ?? 0}</Text>
          </Text>
          <Text size="xs" c="dimmed">
            Sheets <Text span fw={700} c="dark">{summary.SheetCount}</Text>
          </Text>
          <Text size="xs" c="dimmed">
            Formula cells <Text span fw={700} c="dark">{summary.FormulaCellCount}</Text>
          </Text>
        </Group>

        <Button
          variant="subtle"
          size="compact-xs"
          onClick={() => setBreakdownOpen((open) => !open)}
          data-testid="summary-breakdown-toggle"
          aria-expanded={breakdownOpen}
          style={{ flexShrink: 0 }}
        >
          {breakdownOpen ? "Hide breakdown" : "Show breakdown"}
        </Button>
      </Group>

      {breakdownOpen ? (
        <SimpleGrid cols={{ base: 1, sm: 3 }} spacing="xs" mt="xs" data-testid="summary-breakdown">
          <Stack gap={4}>
            <Text fw={600} size="xs">
              Severity
            </Text>
            <RollupChips counts={summary.IssuesBySeverity} />
          </Stack>
          <Stack gap={4}>
            <Text fw={600} size="xs">
              Category
            </Text>
            <RollupChips counts={summary.IssuesByCategory} />
          </Stack>
          <Stack gap={4}>
            <Text fw={600} size="xs">
              Sheet issues
            </Text>
            <RollupChips counts={issuesBySheet} />
          </Stack>
        </SimpleGrid>
      ) : null}
    </Paper>
  );
}
