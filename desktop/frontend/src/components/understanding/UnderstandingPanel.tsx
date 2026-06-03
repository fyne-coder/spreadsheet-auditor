import {
  Alert,
  Badge,
  Button,
  Code,
  Collapse,
  Group,
  Paper,
  ScrollArea,
  SegmentedControl,
  Stack,
  Tabs,
  Text,
  Textarea,
  TextInput,
  Title,
} from "@mantine/core";
import { notifications } from "@mantine/notifications";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  BuildAIHandoff,
  PickJSONSavePath,
  SaveEvidencePacket,
  SavePromptBundle,
  SaveUnderstandingReport,
  ValidateUnderstandingReport,
} from "../../../wailsjs/go/main/AuditService";
import type { model } from "../../../wailsjs/go/models";
import { copyText } from "../../lib/copyText";
import { buildPromptBundleOptions } from "../../lib/promptBundleOptions";
import { understandingReportToMarkdown } from "../../lib/understandingMarkdown";
import { userFacingErrorMessage } from "../../lib/wailsErrors";
import { UnderstandingReportView } from "./UnderstandingReportView";

type PreviewTab = "manifest" | "findings" | "families" | "citations" | "prompt";
type PacketSizePreset = "standard" | "claude" | "chatgpt";

const packetSizeBytes: Record<PacketSizePreset, number | undefined> = {
  standard: undefined,
  claude: 30 * 1024 * 1024,
  chatgpt: 512 * 1024 * 1024,
};

function packetCapError(error: string | null) {
  if (!error?.includes("max_packet_bytes")) {
    return null;
  }
  const match = error.match(/\((\d+) > (\d+)\)/);
  if (!match) {
    return { actualBytes: null, capBytes: null };
  }
  return {
    actualBytes: Number(match[1]),
    capBytes: Number(match[2]),
  };
}

function formatBytes(bytes: number | null) {
  if (!bytes) {
    return "the current cap";
  }
  return `${Math.round(bytes / 1024)} KB`;
}

type Props = {
  workbookPath: string;
  opened: boolean;
  onOpenedChange: (opened: boolean) => void;
};

function formatJSON(value: unknown): string {
  return JSON.stringify(value, null, 2);
}

export function UnderstandingPanel({ workbookPath, opened, onOpenedChange }: Props) {
  const [excludeSheetsRaw, setExcludeSheetsRaw] = useState("");
  const [excludeCellsRaw, setExcludeCellsRaw] = useState("");
  const [userObjectiveRaw, setUserObjectiveRaw] = useState("");
  const [previewTab, setPreviewTab] = useState<PreviewTab>("prompt");
  const [bundle, setBundle] = useState<model.PromptBundleV1 | null>(null);
  const [evidencePacketJSON, setEvidencePacketJSON] = useState("");
  const [promptBundleJSON, setPromptBundleJSON] = useState("");
  const [bundleError, setBundleError] = useState<string | null>(null);
  const [loadingBundle, setLoadingBundle] = useState(false);
  const [packetSizePreset, setPacketSizePreset] =
    useState<PacketSizePreset>("standard");
  const [pasteJSON, setPasteJSON] = useState("");
  const [validation, setValidation] = useState<model.UnderstandingValidationResult | null>(
    null,
  );
  const [validating, setValidating] = useState(false);

  const options = useMemo(
    () =>
      buildPromptBundleOptions(
        excludeSheetsRaw,
        excludeCellsRaw,
        userObjectiveRaw,
        packetSizeBytes[packetSizePreset],
      ),
    [excludeSheetsRaw, excludeCellsRaw, userObjectiveRaw, packetSizePreset],
  );

  const loadBundle = useCallback(async () => {
    const path = workbookPath.trim();
    if (!path) {
      return;
    }
    setLoadingBundle(true);
    setBundleError(null);
    try {
      const handoff = await BuildAIHandoff(path, options);
      setBundle(handoff.bundle ?? null);
      setEvidencePacketJSON(handoff.evidence_packet_json);
      setPromptBundleJSON(handoff.prompt_bundle_json);
    } catch (err) {
      setBundle(null);
      setEvidencePacketJSON("");
      setPromptBundleJSON("");
      setBundleError(userFacingErrorMessage(err));
    } finally {
      setLoadingBundle(false);
    }
  }, [workbookPath, options]);

  useEffect(() => {
    if (!opened) {
      return;
    }
    void loadBundle();
  }, [loadBundle, opened]);

  useEffect(() => {
    setValidation(null);
  }, [pasteJSON, options]);

  const packet = bundle?.evidence_packet;
  const previewContent = useMemo(() => {
    if (!packet) {
      return "";
    }
    switch (previewTab) {
      case "manifest":
        return formatJSON({
          workbook: packet.workbook,
          sheets: packet.sheets,
          audit_hash: packet.audit_hash,
          packet_version: packet.packet_version,
        });
      case "findings":
        return formatJSON(packet.audit_findings ?? []);
      case "families":
        return formatJSON(packet.formula_families ?? []);
      case "citations":
        return formatJSON(packet.citation_map ?? {});
      case "prompt":
        return bundle?.prompt ?? "";
      default:
        return "";
    }
  }, [bundle, packet, previewTab]);

  const notifyCopy = useCallback(async (label: string, value: string) => {
    try {
      await copyText(label, value);
      notifications.show({
        title: "Copied",
        message: `${label} copied to clipboard.`,
        color: "green",
      });
    } catch (err) {
      notifications.show({
        title: "Copy failed",
        message: userFacingErrorMessage(err),
        color: "red",
      });
    }
  }, []);

  const saveJSON = useCallback(
    async (label: string, defaultName: string, save: (outputPath: string) => Promise<void>) => {
      try {
        const outputPath = await PickJSONSavePath(defaultName);
        if (!outputPath) {
          notifications.show({
            title: "Save cancelled",
            message: `${label} save dialog closed.`,
            color: "gray",
          });
          return;
        }
        await save(outputPath);
        notifications.show({
          title: "Saved",
          message: `${label} → ${outputPath}`,
          color: "green",
        });
      } catch (err) {
        notifications.show({
          title: "Save failed",
          message: userFacingErrorMessage(err),
          color: "red",
        });
      }
    },
    [],
  );

  const validatePaste = useCallback(async () => {
    const path = workbookPath.trim();
    if (!path) {
      return;
    }
    setValidating(true);
    try {
      const result = await ValidateUnderstandingReport(path, pasteJSON, options);
      setValidation(result);
    } catch (err) {
      setValidation(null);
      notifications.show({
        title: "Validation failed",
        message: userFacingErrorMessage(err),
        color: "red",
      });
    } finally {
      setValidating(false);
    }
  }, [workbookPath, pasteJSON, options]);

  const citationsResolved =
    validation?.citations_resolved ?? validation?.valid ?? false;
  const rejected =
    validation &&
    (!citationsResolved || validation.parse_error || (validation.rejects?.length ?? 0) > 0);
  const capError = packetCapError(bundleError);
  const acceptedReport = validation && !rejected ? validation.report : null;
  const acceptedSectionCount = acceptedReport
    ? [
        acceptedReport.workbook_purpose,
        acceptedReport.sheet_roles,
        acceptedReport.key_flows,
        acceptedReport.major_risks,
        acceptedReport.cleanup_plan,
        acceptedReport.owner_questions,
        acceptedReport.confidence_notes,
      ].filter((items) => (items?.length ?? 0) > 0).length
    : 0;

  return (
    <Paper
      withBorder
      p="xs"
      radius="md"
      data-testid="understanding-panel"
      style={{ borderColor: "var(--mantine-color-violet-4)" }}
    >
      <Group justify="space-between" mb="xs" wrap="nowrap">
        <Stack gap={2}>
          <Group gap="xs">
            <Title order={4}>Use with your AI assistant</Title>
            <Badge color="violet" variant="light" size="sm">
              Optional
            </Badge>
          </Group>
          <Text size="xs" c="dimmed">
            Copy a review package to paste into ChatGPT, Claude, or Gemini. Then
            attach the workbook and paste the assistant's reply back here to check
            audit citations.
          </Text>
        </Stack>
        <Group gap="xs" wrap="nowrap">
          {opened ? (
            <Button
              variant="light"
              size="compact-sm"
              loading={loadingBundle}
              onClick={() => void loadBundle()}
              data-testid="understanding-refresh"
            >
              Refresh package
            </Button>
          ) : null}
          <Button
            variant={opened ? "default" : "light"}
            size="compact-sm"
            onClick={() => onOpenedChange(!opened)}
            data-testid="understanding-toggle"
          >
            {opened ? "Hide AI handoff" : "Start AI handoff"}
          </Button>
        </Group>
      </Group>

      <Collapse in={opened}>
      {opened ? (
      <>
      <Paper p="xs" radius="sm" bg="gray.0" mb="sm" data-testid="understanding-workflow">
        <Group grow align="flex-start">
          <Stack gap={2}>
            <Text fw={700} size="xs">
              1. Copy package
            </Text>
            <Text size="xs" c="dimmed">
              Copy the prompt after setting your objective.
            </Text>
          </Stack>
          <Stack gap={2}>
            <Text fw={700} size="xs">
              2. Attach workbook
            </Text>
            <Text size="xs" c="dimmed">
              Upload the Excel file and evidence files to your chosen assistant.
            </Text>
          </Stack>
          <Stack gap={2}>
            <Text fw={700} size="xs">
              3. Verify cited evidence
            </Text>
            <Text size="xs" c="dimmed">
              Paste the reply back before relying on it.
            </Text>
          </Stack>
        </Group>
      </Paper>

      <Textarea
        label="Review objective"
        description="Tell the assistant what you want to understand about this workbook."
        placeholder="Example: Help me understand what this inherited budget workbook does, where it depends on external files, and what I should ask the owner before using it."
        minRows={3}
        value={userObjectiveRaw}
        onChange={(event) => setUserObjectiveRaw(event.currentTarget.value)}
        data-testid="user-objective"
        mb="sm"
      />

      <Stack gap="xs" mb="sm">
        <Text fw={600} size="xs">
          Hide sheets or cells from the package
        </Text>
        <Group grow align="flex-start">
          <TextInput
            label="Exclude sheets"
            placeholder="HiddenInputs, Scratch"
            value={excludeSheetsRaw}
            onChange={(event) => setExcludeSheetsRaw(event.currentTarget.value)}
            data-testid="exclude-sheets"
            size="xs"
          />
          <TextInput
            label="Exclude cells"
            placeholder="Model!B1, Inputs!C3"
            value={excludeCellsRaw}
            onChange={(event) => setExcludeCellsRaw(event.currentTarget.value)}
            data-testid="exclude-cells"
            size="xs"
          />
        </Group>
      </Stack>

      <Stack gap="xs" mb="sm">
        <Text fw={600} size="xs" id="packet-size-label">
          Package size limit
        </Text>
        <SegmentedControl
          aria-labelledby="packet-size-label"
          size="xs"
          value={packetSizePreset}
          onChange={(value) => setPacketSizePreset(value as PacketSizePreset)}
          data-testid="packet-size-preset"
          data={[
            { label: "512 KB", value: "standard" },
            { label: "Claude 30 MB", value: "claude" },
            { label: "ChatGPT file cap", value: "chatgpt" },
          ]}
        />
        <Text size="xs" c="dimmed" data-testid="packet-size-guidance">
          Prefer the smaller evidence packet when possible. Claude chat uploads are
          capped at 30 MB per file. ChatGPT hard-caps uploads at 512 MB per file,
          but spreadsheets are typically practical around 50 MB and token or context
          limits may be lower.
        </Text>
      </Stack>

      {bundleError ? (
        <Alert
          color={capError ? "yellow" : "red"}
          title={capError ? "Evidence package is over the current size limit" : "Preview unavailable"}
          mb="sm"
          data-testid="understanding-error"
        >
          {capError ? (
            <Stack gap="xs">
              <Text size="sm">
                This workbook needs about {formatBytes(capError.actualBytes)} of
                evidence, but the current package limit is {formatBytes(capError.capBytes)}.
              </Text>
              <Text size="sm">
                Exclude sheets or cells above, or intentionally raise the limit before
                copying the AI package.
              </Text>
              {packetSizePreset === "standard" ? (
                <Group>
                  <Button
                    size="compact-sm"
                    variant="light"
                    onClick={() => setPacketSizePreset("claude")}
                    data-testid="increase-packet-limit"
                  >
                    Allow Claude 30 MB package
                  </Button>
                </Group>
              ) : null}
            </Stack>
          ) : (
            bundleError
          )}
        </Alert>
      ) : null}

      <Tabs value={previewTab} onChange={(value) => setPreviewTab((value as PreviewTab) ?? "prompt")}>
        <Tabs.List mb="xs" data-testid="understanding-preview-tabs">
          <Tabs.Tab value="prompt">Prompt to copy</Tabs.Tab>
          <Tabs.Tab value="manifest">Manifest</Tabs.Tab>
          <Tabs.Tab value="findings">Findings</Tabs.Tab>
          <Tabs.Tab value="families">Formula families</Tabs.Tab>
          <Tabs.Tab value="citations">Evidence references</Tabs.Tab>
        </Tabs.List>
      </Tabs>

      <ScrollArea h={220} mb="sm" data-testid="understanding-preview">
        <Code block style={{ whiteSpace: "pre-wrap" }}>
          {loadingBundle ? "Loading preview…" : previewContent || "(no preview)"}
        </Code>
      </ScrollArea>

      <Group gap="xs" mb="md" wrap="wrap">
        <Button
          size="compact-sm"
          variant="default"
          disabled={!bundle?.prompt}
          onClick={() => void notifyCopy("Prompt text", bundle?.prompt ?? "")}
          data-testid="copy-prompt"
        >
          Copy package text
        </Button>
        <Button
          size="compact-sm"
          variant="default"
          disabled={!promptBundleJSON}
          onClick={() => void notifyCopy("Prompt bundle JSON", promptBundleJSON)}
          data-testid="copy-prompt-bundle"
        >
          Copy package JSON
        </Button>
        <Button
          size="compact-sm"
          variant="default"
          disabled={!evidencePacketJSON}
          onClick={() => void notifyCopy("Evidence packet JSON", evidencePacketJSON)}
          data-testid="copy-evidence-packet"
        >
          Copy evidence JSON
        </Button>
        <Button
          size="compact-sm"
          variant="light"
          disabled={!workbookPath.trim()}
          onClick={() =>
            void saveJSON("Evidence packet", "evidence-packet.json", (outputPath) =>
              SaveEvidencePacket(workbookPath.trim(), outputPath, options),
            )
          }
          data-testid="save-evidence-packet"
        >
          Save evidence JSON
        </Button>
        <Button
          size="compact-sm"
          variant="light"
          disabled={!workbookPath.trim()}
          onClick={() =>
            void saveJSON("Prompt bundle", "prompt-bundle.json", (outputPath) =>
              SavePromptBundle(workbookPath.trim(), outputPath, options),
            )
          }
          data-testid="save-prompt-bundle"
        >
          Save package JSON
        </Button>
      </Group>

      <Stack gap="xs" data-testid="understanding-pasteback">
        <Title order={5}>Paste your AI assistant's reply</Title>
        <Textarea
          placeholder="Paste the JSON your AI assistant returned"
          minRows={6}
          value={pasteJSON}
          onChange={(event) => setPasteJSON(event.currentTarget.value)}
          data-testid="understanding-paste-input"
        />
        <Group>
          <Button
            size="compact-sm"
            loading={validating}
            disabled={!pasteJSON.trim()}
            onClick={() => void validatePaste()}
          data-testid="understanding-validate"
        >
            Verify cited evidence
          </Button>
        </Group>

        {validation ? (
          rejected ? (
            <Alert
              color="red"
              title="Couldn't verify — see what didn't match"
              data-testid="understanding-rejected"
            >
              {validation.parse_error ? (
                <Text size="sm">{validation.parse_error}</Text>
              ) : null}
              {(validation.rejects ?? []).map((reject, index) => (
                <Text key={`${reject.field}-${index}`} size="sm">
                  {reject.field}[{reject.index}]: {reject.reason}
                  {reject.citation ? ` (${reject.citation})` : ""}
                </Text>
              ))}
            </Alert>
          ) : acceptedReport ? (
            <Stack gap="xs">
              <Alert
                color="green"
                title="Cited evidence verified"
                data-testid="understanding-accepted"
              >
                <Text size="sm">
                  Every cited ID resolves to the evidence packet. Review the AI&apos;s claims before
                  relying on them. {acceptedSectionCount} cited sections parsed with
                  resolvable IDs; use this as explanation alongside the deterministic
                  audit, not as recalculated workbook results.
                </Text>
              </Alert>
              <Group gap="xs" wrap="wrap">
                <Button
                  size="compact-sm"
                  variant="light"
                  onClick={() =>
                    void saveJSON("Verified AI analysis", "verified-ai-analysis.json", (outputPath) =>
                      SaveUnderstandingReport(
                        workbookPath.trim(),
                        pasteJSON,
                        outputPath,
                        options,
                      ),
                    )
                  }
                  data-testid="save-understanding-report"
                >
                  Save verified analysis
                </Button>
                <Button
                  size="compact-sm"
                  variant="default"
                  onClick={() =>
                    void notifyCopy(
                      "Verified AI analysis Markdown",
                      understandingReportToMarkdown(acceptedReport),
                    )
                  }
                  data-testid="copy-understanding-report"
                >
                  Copy as Markdown
                </Button>
              </Group>
              <Paper withBorder p="sm" radius="md" bg="var(--mantine-color-violet-0)">
                <UnderstandingReportView report={acceptedReport} />
              </Paper>
            </Stack>
          ) : null
        ) : null}
      </Stack>
      </>
      ) : null}
      </Collapse>
    </Paper>
  );
}
