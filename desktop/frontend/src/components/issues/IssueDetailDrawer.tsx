import { Badge, Button, Code, Collapse, Drawer, Group, Paper, Stack, Text, Title } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useState } from "react";
import type { model } from "../../../wailsjs/go/models";
import { copyText } from "../../lib/copyText";
import { issueKey } from "../../lib/issueKey";
import { ownerQuestion } from "./ownerQuestions";

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
  const [advancedOpen, setAdvancedOpen] = useState(false);
  if (!issue) {
    return null;
  }

  const key = issueKey(issue);
  const formula = issue.Evidence?.Formula ?? "";
  const detailsJson = JSON.stringify(issue.Details ?? {}, null, 2);
  const location = `${issue.Evidence?.Sheet ?? "(unknown)"}!${issue.Evidence?.Cell ?? "(unknown)"}`;

  return (
    <Drawer opened={opened} onClose={onClose} title="Issue" position="right" size="md">
      <Stack gap="sm">
        <Title order={4}>{issue.Title}</Title>
        <Group gap="xs">
          {issue.Priority ? <Badge variant="filled">{issue.Priority}</Badge> : null}
          <Badge variant="light">{issue.Severity}</Badge>
          <Text size="sm" c="dimmed">
            {location}
          </Text>
        </Group>

        {issue.ImpactFactors?.length ? (
          <Stack gap={4}>
            <Text size="sm" fw={700}>
              Why this is prioritized
            </Text>
            {issue.ImpactFactors.map((factor) => (
              <Text key={factor.Code} size="sm">
                <strong>{factor.Code}:</strong> {factor.Explanation}
              </Text>
            ))}
          </Stack>
        ) : null}

        <Stack gap={4}>
          <Text size="sm" fw={700}>
            Why this matters
          </Text>
          <Text size="sm">{issue.Rationale || issue.Message}</Text>
        </Stack>

        <Stack gap={4}>
          <Text size="sm" fw={700}>
            Suggested fix
          </Text>
          <Text size="sm">{issue.Remediation}</Text>
        </Stack>

        <Paper withBorder p="xs" radius="sm">
          <Stack gap={4}>
            <Text size="sm" fw={700}>
              Where it is
            </Text>
            <Text size="sm">
              <strong>Category:</strong> {issue.Category}
            </Text>
            <Text size="sm">
              <strong>Rule:</strong> {issue.RuleID}
            </Text>
            {formula ? (
              <>
                <Text size="sm" fw={600}>
                  Formula
                </Text>
                <Code block>{formula}</Code>
              </>
            ) : null}
          </Stack>
        </Paper>

        <Stack gap={4}>
          <Text size="sm" fw={700}>
            Ask the workbook owner
          </Text>
          <Text size="sm">{ownerQuestion(issue)}</Text>
        </Stack>

        <Button
          size="xs"
          variant="default"
          onClick={() => setAdvancedOpen((open) => !open)}
          data-testid="issue-advanced-toggle"
        >
          {advancedOpen ? "Hide advanced details" : "Show advanced details"}
        </Button>
        <Collapse in={advancedOpen}>
          {advancedOpen ? (
          <Stack gap="sm" data-testid="issue-advanced-details">
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
              <Button size="xs" variant="light" onClick={() => copyField("Cell", location)}>
                Copy cell
              </Button>
              {formula ? (
                <Button size="xs" variant="light" onClick={() => copyField("Formula", formula)}>
                  Copy formula
                </Button>
              ) : null}
            </Group>
            <Text size="sm" fw={600}>
              Details JSON
            </Text>
            <Code block>{detailsJson}</Code>
          </Stack>
          ) : null}
        </Collapse>
      </Stack>
    </Drawer>
  );
}
