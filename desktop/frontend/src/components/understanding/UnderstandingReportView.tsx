import { Badge, List, Stack, Text, Title } from "@mantine/core";
import type { ReactNode } from "react";
import type { model } from "../../../wailsjs/go/models";

type Props = {
  report: model.UnderstandingReportV1;
};

function ClaimList({
  title,
  items,
  renderItem,
}: {
  title: string;
  items: unknown[];
  renderItem: (item: unknown, index: number) => ReactNode;
}) {
  if (!items || items.length === 0) {
    return null;
  }
  return (
    <Stack gap={4}>
      <Title order={6}>{title}</Title>
      <List size="sm" spacing="xs">
        {items.map((item, index) => (
          <List.Item key={`${title}-${index}`}>{renderItem(item, index)}</List.Item>
        ))}
      </List>
    </Stack>
  );
}

function CitationBadges({ citations }: { citations: string[] | undefined }) {
  if (!citations || citations.length === 0) {
    return (
      <Text size="xs" c="dimmed">
        (no citations)
      </Text>
    );
  }
  return (
    <Text size="xs" c="dimmed" component="span">
      {citations.map((citation) => (
        <Badge key={citation} variant="outline" size="xs" mr={4} mb={4}>
          {citation}
        </Badge>
      ))}
    </Text>
  );
}

export function UnderstandingReportView({ report }: Props) {
  return (
    <Stack gap="sm" data-testid="understanding-report-view">
      <ClaimList
        title="Workbook purpose"
        items={report.workbook_purpose ?? []}
        renderItem={(item) => {
          const claim = item as model.UnderstandingClaim;
          return (
            <Stack gap={2}>
              <Text size="sm">{claim.claim}</Text>
              <CitationBadges citations={claim.citations} />
            </Stack>
          );
        }}
      />
      <ClaimList
        title="Sheet roles"
        items={report.sheet_roles ?? []}
        renderItem={(item) => {
          const role = item as model.SheetRoleClaim;
          return (
            <Stack gap={2}>
              <Text size="sm">
                <Text span fw={600}>
                  {role.sheet}
                </Text>
                : {role.role}
              </Text>
              <CitationBadges citations={role.citations} />
            </Stack>
          );
        }}
      />
      <ClaimList
        title="Key flows"
        items={report.key_flows ?? []}
        renderItem={(item) => {
          const flow = item as model.FlowClaim;
          return (
            <Stack gap={2}>
              <Text size="sm">{flow.summary}</Text>
              <CitationBadges citations={flow.citations} />
            </Stack>
          );
        }}
      />
      <ClaimList
        title="Major risks"
        items={report.major_risks ?? []}
        renderItem={(item) => {
          const risk = item as model.RiskClaim;
          return (
            <Stack gap={2}>
              <Text size="sm">
                <Badge size="xs" variant="light" mr={6}>
                  {risk.severity}
                </Badge>
                {risk.summary}
              </Text>
              <CitationBadges citations={risk.citations} />
            </Stack>
          );
        }}
      />
      <ClaimList
        title="Cleanup plan"
        items={report.cleanup_plan ?? []}
        renderItem={(item) => {
          const action = item as model.CleanupAction;
          return (
            <Stack gap={2}>
              <Text size="sm">{action.action}</Text>
              <CitationBadges citations={action.citations} />
            </Stack>
          );
        }}
      />
      <ClaimList
        title="Owner questions"
        items={report.owner_questions ?? []}
        renderItem={(item) => {
          const question = item as model.OwnerQuestion;
          return (
            <Stack gap={2}>
              <Text size="sm">{question.question}</Text>
              <CitationBadges citations={question.context_citations} />
            </Stack>
          );
        }}
      />
      <ClaimList
        title="Confidence notes"
        items={report.confidence_notes ?? []}
        renderItem={(item) => {
          const note = item as model.ConfidenceNote;
          return (
            <Stack gap={2}>
              <Text size="sm">{note.note}</Text>
              <CitationBadges citations={note.citations} />
            </Stack>
          );
        }}
      />
    </Stack>
  );
}
