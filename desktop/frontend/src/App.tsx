import { Alert, Button, Group, Paper, Stack, Text, TextInput, Title } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import type { RowSelectionState } from "@tanstack/react-table";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  PickExportSavePath,
  PickWorkbook,
  SaveExport,
  ScanWorkbook,
} from "../wailsjs/go/main/AuditService";
import { model } from "../wailsjs/go/models";
import { ExportModal } from "./components/ExportModal";
import { EmptyState } from "./components/EmptyState";
import { IssuesTable } from "./components/issues/IssuesTable";
import { SummaryPanel } from "./components/SummaryPanel";
import {
  defaultExportFilename,
  type ExportFormat,
  type ExportScope,
} from "./lib/exportDefaults";
import { userFacingErrorMessage } from "./lib/wailsErrors";

type StatusKind = "idle" | "info" | "success" | "error";

export default function App() {
  const [workbookPath, setWorkbookPath] = useState("");
  const [report, setReport] = useState<model.AuditReport | null>(null);
  const [scanning, setScanning] = useState(false);
  const [status, setStatus] = useState<{ kind: StatusKind; message: string }>({
    kind: "idle",
    message: "Ready",
  });
  const [exportOpen, setExportOpen] = useState(false);
  const [exportFormat, setExportFormat] = useState<ExportFormat>("html");
  const [exportScope, setExportScope] = useState<ExportScope>("all");
  const [exportConfirming, setExportConfirming] = useState(false);
  const [exportedAtPreview, setExportedAtPreview] = useState<string | null>(null);
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});

  const setInfo = (message: string) => setStatus({ kind: "info", message });
  const setSuccess = (message: string) => setStatus({ kind: "success", message });
  const setError = (message: string) => setStatus({ kind: "error", message });

  const selectedIssueIds = useMemo(
    () => Object.keys(rowSelection).filter((id) => rowSelection[id]),
    [rowSelection],
  );
  const selectedCount = selectedIssueIds.length;
  const totalIssueCount = report?.Issues?.length ?? 0;

  const browseWorkbook = useCallback(async () => {
    try {
      const path = await PickWorkbook();
      if (path) {
        setWorkbookPath(path);
        setInfo("Workbook selected.");
      }
    } catch (err) {
      setError(userFacingErrorMessage(err));
    }
  }, []);

  const scanWorkbook = useCallback(async () => {
    const path = workbookPath.trim();
    if (!path) {
      setError("Enter or browse to a workbook path first.");
      return;
    }
    setScanning(true);
    setInfo("Scanning workbook...");
    try {
      const next = await ScanWorkbook(path);
      setReport(next);
      setRowSelection({});
      setSuccess(`Complete · ${next.Summary.IssueCount} issues found`);
    } catch (err) {
      setReport(null);
      setRowSelection({});
      setError(userFacingErrorMessage(err));
    } finally {
      setScanning(false);
    }
  }, [workbookPath]);

  const openExportModal = useCallback(() => {
    const path = workbookPath.trim();
    if (!path) {
      setError("Scan a workbook before exporting.");
      notifications.show({
        title: "Export unavailable",
        message: "Select a workbook path first.",
        color: "red",
      });
      return;
    }
    if (!report) {
      setError("Run a scan before exporting.");
      notifications.show({
        title: "Export unavailable",
        message: "Scan the workbook to load issues.",
        color: "red",
      });
      return;
    }
    setExportFormat("html");
    setExportScope("all");
    setExportedAtPreview(null);
    setExportOpen(true);
  }, [workbookPath, report]);

  const confirmExport = useCallback(async () => {
    const path = workbookPath.trim();
    if (!path || !report) {
      setError("Run a scan before exporting.");
      return;
    }
    if (exportScope === "selected" && selectedCount === 0) {
      setError("Select at least one issue to export a selection.");
      return;
    }

    const exportedAtRFC3339 = new Date().toISOString();
    setExportedAtPreview(exportedAtRFC3339);

    const issueIDs = exportScope === "selected" ? selectedIssueIds : [];
    const defaultName = defaultExportFilename(exportFormat);

    setExportConfirming(true);
    try {
      const outputPath = await PickExportSavePath(exportFormat, defaultName);
      if (!outputPath) {
        setInfo("Export cancelled — save dialog closed.");
        notifications.show({
          title: "Export cancelled",
          message: "Save dialog was closed without choosing a file.",
          color: "gray",
        });
        return;
      }
      await SaveExport(path, outputPath, exportFormat, exportedAtRFC3339, issueIDs);
      const successMessage = `Exported at ${exportedAtRFC3339} → ${outputPath}`;
      setSuccess(successMessage);
      notifications.show({
        title: "Export complete",
        message: successMessage,
        color: "green",
      });
      setExportOpen(false);
      setExportedAtPreview(null);
    } catch (err) {
      const message = userFacingErrorMessage(err);
      setError(message);
      notifications.show({
        title: "Export failed",
        message,
        color: "red",
      });
    } finally {
      setExportConfirming(false);
    }
  }, [
    workbookPath,
    report,
    exportScope,
    selectedCount,
    selectedIssueIds,
    exportFormat,
  ]);

  useEffect(() => {
    if (exportScope === "selected" && selectedCount === 0) {
      setExportScope("all");
    }
  }, [exportScope, selectedCount]);

  useEffect(() => {
    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "o") {
        event.preventDefault();
        void browseWorkbook();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [browseWorkbook]);

  const statusColor =
    status.kind === "error" ? "red" : status.kind === "success" ? "green" : "dimmed";

  return (
    <Stack maw={1240} miw={1100} mx="auto" p="xs" gap="xs">
      <Group justify="space-between" align="flex-start">
        <Stack gap={4}>
          <Title order={3}>Spreadsheet Auditor</Title>
          <Text c="dimmed" size="xs">
            Local read-only workbook triage
          </Text>
        </Stack>
        <Text size="sm" c={statusColor} data-testid="status" role="status">
          {status.message}
        </Text>
      </Group>

      <Paper withBorder p="xs" radius="md">
        <Text fw={600} size="xs" mb={4}>
          Workbook path
        </Text>
        <Group align="flex-end" wrap="nowrap">
          <TextInput
            flex={1}
            placeholder="Select or enter a .xlsx or .xlsm file"
            value={workbookPath}
            onChange={(e) => setWorkbookPath(e.currentTarget.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                void scanWorkbook();
              }
            }}
            data-testid="workbook-path"
            size="xs"
          />
          <Button variant="default" size="compact-sm" onClick={() => void browseWorkbook()}>
            Browse
          </Button>
          <Button
            loading={scanning}
            size="compact-sm"
            onClick={() => void scanWorkbook()}
            data-testid="scan-btn"
          >
            {report ? "Rescan" : "Scan"}
          </Button>
        </Group>
      </Paper>

      {!report ? (
        <>
          <EmptyState />
          {status.kind === "error" ? (
            <Alert color="red" title="Scan failed" data-testid="scan-error">
              {status.message}
            </Alert>
          ) : null}
        </>
      ) : (
        <>
          <SummaryPanel report={report} />
          <Paper withBorder p="xs" radius="md">
            <Group justify="space-between" mb="xs">
              <Title order={4}>Issues</Title>
              <Button
                variant="light"
                size="compact-sm"
                onClick={openExportModal}
                data-testid="export-btn"
              >
                Export review pack
              </Button>
            </Group>
            <IssuesTable
              issues={report.Issues ?? []}
              rowSelection={rowSelection}
              onRowSelectionChange={setRowSelection}
            />
          </Paper>
        </>
      )}

      <ExportModal
        opened={exportOpen}
        onClose={() => {
          if (!exportConfirming) {
            setExportOpen(false);
            setExportedAtPreview(null);
          }
        }}
        format={exportFormat}
        onFormatChange={setExportFormat}
        scope={exportScope}
        onScopeChange={setExportScope}
        totalIssueCount={totalIssueCount}
        selectedCount={selectedCount}
        exportedAtPreview={exportedAtPreview}
        confirming={exportConfirming}
        onConfirm={() => void confirmExport()}
      />
    </Stack>
  );
}
