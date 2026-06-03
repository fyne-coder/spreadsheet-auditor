import { Badge, List, Paper, Stack, Text, Title } from "@mantine/core";

export function EmptyState() {
  return (
    <Paper withBorder p="xl" radius="md" data-testid="empty-state">
      <Stack gap="md">
        <Title order={2}>Choose a workbook to review</Title>
        <Text c="dimmed">
          The scan reads the file only. It does not run macros, refresh links, or
          recalculate formulas.
        </Text>
        <List size="sm" c="dimmed" spacing={4}>
          <List.Item>A plain overview of what was found.</List.Item>
          <List.Item>A list of issues to fix or ask about.</List.Item>
          <List.Item>An optional package to send to your AI assistant.</List.Item>
        </List>
        <Stack gap="xs">
          <Badge variant="light">Read-only</Badge>
          <Badge variant="light">.xlsx / .xlsm</Badge>
          <Badge variant="light">Deterministic issues</Badge>
        </Stack>
      </Stack>
    </Paper>
  );
}
