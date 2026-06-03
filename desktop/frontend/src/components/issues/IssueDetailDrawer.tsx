import { Button, Code, Drawer, Group, Stack, Text, Title } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import type { model } from "../../../wailsjs/go/models";
import { copyText } from "../../lib/copyText";
import { issueKey } from "../../lib/issueKey";

type Props = {
  issue: model.Issue | null;
  opened: boolean;
  onClose: () => void;
};

async function copyField(label: string, value: string) {
  try {
    await copyText(label, value);
    notifications.show({ message: `${label} copied`, color: "green" });
  } catch (err) {
    notifications.show({
      message: err instanceof Error ? err.message : String(err),
      color: "red",
    });
  }
}

export function IssueDetailDrawer({ issue, opened, onClose }: Props) {
  if (!issue) {
    return null;
  }

  const key = issueKey(issue);
  const formula = issue.Evidence?.Formula ?? "";
  const detailsJson = JSON.stringify(issue.Details ?? {}, null, 2);

  return (
    <Drawer opened={opened} onClose={onClose} title="Issue detail" position="right" size="md">
      <Stack gap="sm">
        <Title order={4}>{issue.Title}</Title>
        <Text size="sm" c="dimmed">
          Issue key: <Code>{key}</Code>
        </Text>
        <Group gap="xs">
          <Button size="xs" variant="light" onClick={() => copyField("Issue key", key)}>
            Copy issue key
          </Button>
          <Button size="xs" variant="light" onClick={() => copyField("Rule", issue.RuleID)}>
            Copy rule
          </Button>
          <Button
            size="xs"
            variant="light"
            onClick={() => copyField("Cell", `${issue.Evidence?.Sheet}!${issue.Evidence?.Cell}`)}
          >
            Copy cell
          </Button>
          {formula ? (
            <Button size="xs" variant="light" onClick={() => copyField("Formula", formula)}>
              Copy formula
            </Button>
          ) : null}
        </Group>
        <Text size="sm">
          <strong>Severity:</strong> {issue.Severity}
        </Text>
        <Text size="sm">
          <strong>Category:</strong> {issue.Category}
        </Text>
        <Text size="sm">
          <strong>Message:</strong> {issue.Message}
        </Text>
        <Text size="sm">
          <strong>Rationale:</strong> {issue.Rationale}
        </Text>
        <Text size="sm">
          <strong>Remediation:</strong> {issue.Remediation}
        </Text>
        {formula ? (
          <Stack gap={4}>
            <Text size="sm" fw={600}>
              Formula
            </Text>
            <Code block>{formula}</Code>
          </Stack>
        ) : null}
        <Text size="sm" fw={600}>
          Details JSON
        </Text>
        <Code block>{detailsJson}</Code>
      </Stack>
    </Drawer>
  );
}
