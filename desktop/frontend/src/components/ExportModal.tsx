import {
  Button,
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
  format: ExportFormat;
  onFormatChange: (format: ExportFormat) => void;
  scope: ExportScope;
  onScopeChange: (scope: ExportScope) => void;
  totalIssueCount: number;
  selectedCount: number;
  exportedAtPreview: string | null;
  confirming: boolean;
  onConfirm: () => void;
};

export function ExportModal({
  opened,
  onClose,
  format,
  onFormatChange,
  scope,
  onScopeChange,
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
      title="Export review pack"
      data-testid="export-modal"
      centered
      transitionProps={{ duration: 0 }}
    >
      <Stack gap="md">
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
            Output filename
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
              Select rows in the table to export a subset. Filtered rows are not
              exported unless you select them.
            </Text>
          ) : null}
        </div>

        <div>
          <Text size="sm" fw={500}>
            Exported at
          </Text>
          <Text size="sm" c="dimmed" data-testid="export-exported-at">
            {exportedAtPreview ??
              "Recorded in RFC3339 when you confirm export (before the save dialog)."}
          </Text>
        </div>

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
