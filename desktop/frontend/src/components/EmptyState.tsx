import { Badge, Paper, Stack, Text, Title } from "@mantine/core";

export function EmptyState() {
  return (
    <Paper withBorder p="xl" radius="md" data-testid="empty-state">
      <Stack gap="md">
        <Title order={2}>Choose a workbook to review</Title>
        <Text c="dimmed">
          Static scan only. The analyzer reads workbook structure and formulas without
          executing macros, refreshing links, or evaluating formulas.
        </Text>
        <Stack gap="xs">
          <Badge variant="light">Read-only</Badge>
          <Badge variant="light">.xlsx / .xlsm</Badge>
          <Badge variant="light">Deterministic issues</Badge>
        </Stack>
      </Stack>
    </Paper>
  );
}
