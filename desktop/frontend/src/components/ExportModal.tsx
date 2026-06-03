import {
  Button,
  Checkbox,
  Group,
  Modal,
  Radio,
  SegmentedControl,
  Stack,
  Text,
} from "@mantine/core";
import {
  defaultExportFilename,
  formatSegmentLabel,
  type ExportFormat,
  type ExportScope,
} from "../lib/exportDefaults";

type Props = {
  opened: boolean;
  onClose: () => void;
  job: ExportJob;
  onJobChange: (job: ExportJob) => void;
  format: ExportFormat;
  onFormatChange: (format: ExportFormat) => void;
  scope: ExportScope;
  onScopeChange: (scope: ExportScope) => void;
  includeFullPath: boolean;
  onIncludeFullPathChange: (includeFullPath: boolean) => void;
  totalIssueCount: number;
  selectedCount: number;
  exportedAtPreview: string | null;
  confirming: boolean;
  onConfirm: () => void;
};

export type ExportJob = "owner-summary" | "detailed-audit" | "issue-list";

const exportJobs: Array<{
  value: ExportJob;
  title: string;
  description: string;
}> = [
  {
    value: "owner-summary",
    title: "Owner summary",
    description: "A readable HTML handoff for the workbook owner.",
  },
  {
    value: "detailed-audit",
    title: "Detailed audit",
    description: "A fuller HTML report for follow-up review.",
  },
  {
    value: "issue-list",
    title: "Issue list",
    description: "A CSV table for sorting, tracking, or cleanup.",
  },
];

export function ExportModal({
  opened,
  onClose,
  job,
  onJobChange,
  format,
  onFormatChange,
  scope,
  onScopeChange,
  includeFullPath,
  onIncludeFullPathChange,
  totalIssueCount,
  selectedCount,
  exportedAtPreview,
  confirming,
  onConfirm,
}: Props) {
  const defaultName = defaultExportFilename(format);
  const selectedScopeDisabled = selectedCount === 0;

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      title="Export results"
      data-testid="export-modal"
      centered
      transitionProps={{ duration: 0 }}
    >
      <Stack gap="md">
        <div>
          <Text size="sm" c="dimmed" mb="sm">
            Choose the handoff you need, then confirm the file format and issue scope.
            AI evidence packages stay in the optional AI handoff section.
          </Text>
        </div>

        <div>
          <Text size="sm" fw={500} mb={6} id="export-job-label">
            Export job
          </Text>
          <Radio.Group
            aria-labelledby="export-job-label"
            value={job}
            onChange={(value) => onJobChange(value as ExportJob)}
            data-testid="export-job"
          >
            <Stack gap="xs">
              {exportJobs.map((option) => (
                <Radio
                  key={option.value}
                  value={option.value}
                  data-testid={`export-job-${option.value}`}
                  aria-label={option.title}
                  label={
                    <Stack gap={0}>
                      <Text size="sm">{option.title}</Text>
                      <Text size="xs" c="dimmed">
                        {option.description}
                      </Text>
                    </Stack>
                  }
                />
              ))}
            </Stack>
          </Radio.Group>
        </div>

        <div>
          <Text size="sm" fw={500} mb={6} id="export-format-label">
            Format
          </Text>
          <SegmentedControl
            aria-labelledby="export-format-label"
            data-testid="export-format"
            value={format}
            onChange={(value) => onFormatChange(value as ExportFormat)}
            data={[
              { label: formatSegmentLabel("html"), value: "html" },
              { label: formatSegmentLabel("csv"), value: "csv" },
            ]}
          />
        </div>

        <div>
          <Text size="sm" fw={500}>
            Saved as
          </Text>
          <Text size="sm" c="dimmed" data-testid="export-filename">
            {defaultName}
          </Text>
        </div>

        <div>
          <Text size="sm" fw={500} mb={6} id="export-scope-label">
            Export scope
          </Text>
          <Radio.Group
            aria-labelledby="export-scope-label"
            value={scope}
            onChange={(value) => onScopeChange(value as ExportScope)}
            data-testid="export-scope"
          >
            <Stack gap="xs">
              <Radio
                value="all"
                label={`All issues (${totalIssueCount})`}
                data-testid="export-scope-all"
              />
              <Radio
                value="selected"
                label={`Selected issues (${selectedCount})`}
                disabled={selectedScopeDisabled}
                data-testid="export-scope-selected"
              />
            </Stack>
          </Radio.Group>
          {selectedScopeDisabled ? (
              <Text size="xs" c="dimmed" mt={4}>
              To export a subset, tick the rows you want in the table. Filtered rows
              are not exported unless you tick them.
            </Text>
          ) : null}
        </div>

        <div>
          <Text size="sm" fw={500}>
            Exported at
          </Text>
          <Text size="sm" c="dimmed" data-testid="export-exported-at">
            {exportedAtPreview ?? "A timestamp is added when you confirm export."}
          </Text>
        </div>

        <Checkbox
          checked={includeFullPath}
          onChange={(event) => onIncludeFullPathChange(event.currentTarget.checked)}
          label="Include full workbook path in the export"
          description="Exports use the workbook filename only by default so local folder names stay private."
          data-testid="export-include-full-path"
        />

        <Group justify="flex-end" mt="sm">
          <Button
            variant="default"
            onClick={onClose}
            disabled={confirming}
            data-testid="export-cancel"
          >
            Cancel
          </Button>
          <Button
            onClick={onConfirm}
            loading={confirming}
            data-testid="export-confirm"
          >
            Export
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
