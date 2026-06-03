import { Alert, Button, Group, Paper, Stack, Text, TextInput, Title } from "@mantine/core";
import { notifications } from "@mantine/notifications";
import type { ColumnFiltersState, RowSelectionState } from "@tanstack/react-table";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  PickExportSavePath,
  PickWorkbook,
  SaveExport,
  ScanWorkbook,
} from "../wailsjs/go/main/AuditService";
import { model } from "../wailsjs/go/models";
import { ExportModal } from "./components/ExportModal";
import type { ExportJob } from "./components/ExportModal";
import { EmptyState } from "./components/EmptyState";
import { IssuesTable } from "./components/issues/IssuesTable";
import { OverviewPanel } from "./components/overview/OverviewPanel";
import { SummaryPanel } from "./components/SummaryPanel";
import { UnderstandingPanel } from "./components/understanding/UnderstandingPanel";
import {
  defaultExportFilename,
  type ExportFormat,
  type ExportScope,
} from "./lib/exportDefaults";
import { userFacingErrorMessage } from "./lib/wailsErrors";

type StatusKind = "idle" | "info" | "success" | "error";

export default function App() {
  const [workbookPath, setWorkbookPath] = useState("");
  const [scannedWorkbookPath, setScannedWorkbookPath] = useState<string | null>(null);
  const [report, setReport] = useState<model.AuditReport | null>(null);
  const [scanning, setScanning] = useState(false);
  const [status, setStatus] = useState<{ kind: StatusKind; message: string }>({
    kind: "idle",
    message: "Ready",
  });
  const [exportOpen, setExportOpen] = useState(false);
  const [exportJob, setExportJob] = useState<ExportJob>("owner-summary");
  const [exportFormat, setExportFormat] = useState<ExportFormat>("html");
  const [exportScope, setExportScope] = useState<ExportScope>("all");
  const [exportIncludeFullPath, setExportIncludeFullPath] = useState(false);
  const [exportConfirming, setExportConfirming] = useState(false);
  const [exportedAtPreview, setExportedAtPreview] = useState<string | null>(null);
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [aiHandoffOpen, setAIHandoffOpen] = useState(false);
  const issuesSectionRef = useRef<HTMLDivElement | null>(null);

  const setInfo = (message: string) => setStatus({ kind: "info", message });
  const setSuccess = (message: string) => setStatus({ kind: "success", message });
  const setError = (message: string) => setStatus({ kind: "error", message });

  const selectedIssueIds = useMemo(
    () => Object.keys(rowSelection).filter((id) => rowSelection[id]),
    [rowSelection],
  );
  const selectedCount = selectedIssueIds.length;
  const totalIssueCount = report?.Issues?.length ?? 0;
  const normalizedWorkbookPath = workbookPath.trim();
  const reportPathChanged =
    Boolean(report) &&
    Boolean(scannedWorkbookPath) &&
    normalizedWorkbookPath !== scannedWorkbookPath;

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
      setScannedWorkbookPath(path);
      setRowSelection({});
      setColumnFilters([]);
      setAIHandoffOpen(false);
      setSuccess(`Scan complete — ${next.Summary.IssueCount} issues to review`);
    } catch (err) {
      setReport(null);
      setScannedWorkbookPath(null);
      setRowSelection({});
      setColumnFilters([]);
      setAIHandoffOpen(false);
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
    if (reportPathChanged) {
      const message = "Workbook path changed since the last scan. Rescan before exporting.";
      setError(message);
      notifications.show({
        title: "Export unavailable",
        message,
        color: "red",
      });
      return;
    }
    setExportJob("owner-summary");
    setExportFormat("html");
    setExportScope("all");
    setExportIncludeFullPath(false);
    setExportedAtPreview(null);
    setExportOpen(true);
  }, [workbookPath, report, reportPathChanged]);

  const confirmExport = useCallback(async () => {
    const path = scannedWorkbookPath ?? workbookPath.trim();
    if (!path || !report) {
      setError("Run a scan before exporting.");
      return;
    }
    if (reportPathChanged) {
      setError("Workbook path changed since the last scan. Rescan before exporting.");
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
      await SaveExport(
        path,
        outputPath,
        exportFormat,
        exportedAtRFC3339,
        issueIDs,
        exportIncludeFullPath,
      );
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
    scannedWorkbookPath,
    report,
    reportPathChanged,
    exportScope,
    selectedCount,
    selectedIssueIds,
    exportFormat,
    exportIncludeFullPath,
  ]);

  useEffect(() => {
    if (exportScope === "selected" && selectedCount === 0) {
      setExportScope("all");
    }
  }, [exportScope, selectedCount]);

  const changeExportJob = useCallback((job: ExportJob) => {
    setExportJob(job);
    setExportFormat(job === "issue-list" ? "csv" : "html");
  }, []);

  const applyIssueFilters = useCallback((filters: ColumnFiltersState) => {
    setColumnFilters(filters);
    requestAnimationFrame(() => {
      issuesSectionRef.current?.scrollIntoView({ block: "start", behavior: "smooth" });
    });
  }, []);

  const openAIHandoff = useCallback(() => {
    if (reportPathChanged) {
      const message = "Workbook path changed since the last scan. Rescan before building an AI package.";
      setError(message);
      notifications.show({
        title: "AI package unavailable",
        message,
        color: "red",
      });
      return;
    }
    setAIHandoffOpen(true);
    requestAnimationFrame(() => {
      document
        .querySelector<HTMLElement>('[data-testid="understanding-panel"]')
        ?.scrollIntoView({ block: "start", behavior: "smooth" });
    });
  }, [reportPathChanged]);

  const changeAIHandoffOpen = useCallback(
    (opened: boolean) => {
      if (!opened) {
        setAIHandoffOpen(false);
        return;
      }
      openAIHandoff();
    },
    [openAIHandoff],
  );

  useEffect(() => {
    if (!reportPathChanged) {
      return;
    }
    setAIHandoffOpen(false);
    setExportOpen(false);
    setExportedAtPreview(null);
  }, [reportPathChanged]);

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
            Local, read-only review of inherited spreadsheets
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
        {reportPathChanged ? (
          <Alert color="yellow" mt="xs" title="Rescan required" data-testid="path-dirty-alert">
            The visible results are for <Text span fw={700}>{scannedWorkbookPath}</Text>.
            Rescan before exporting or building an AI package from the changed path.
          </Alert>
        ) : null}
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
          <OverviewPanel
            report={report}
            onApplyFilters={applyIssueFilters}
            onStartAIHandoff={openAIHandoff}
          />
          <Paper withBorder p="xs" radius="md" ref={issuesSectionRef}>
            <Group justify="space-between" mb="xs">
              <Title order={4}>Issues</Title>
              <Button
                variant="light"
                size="compact-sm"
                onClick={openExportModal}
                data-testid="export-btn"
              >
                Export results
              </Button>
            </Group>
            <IssuesTable
              issues={report.Issues ?? []}
              columnFilters={columnFilters}
              onColumnFiltersChange={setColumnFilters}
              rowSelection={rowSelection}
              onRowSelectionChange={setRowSelection}
            />
          </Paper>
          <UnderstandingPanel
            workbookPath={scannedWorkbookPath ?? workbookPath}
            opened={aiHandoffOpen}
            onOpenedChange={changeAIHandoffOpen}
          />
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
        job={exportJob}
        onJobChange={changeExportJob}
        format={exportFormat}
        onFormatChange={setExportFormat}
        scope={exportScope}
        onScopeChange={setExportScope}
        includeFullPath={exportIncludeFullPath}
        onIncludeFullPathChange={setExportIncludeFullPath}
        totalIssueCount={totalIssueCount}
        selectedCount={selectedCount}
        exportedAtPreview={exportedAtPreview}
        confirming={exportConfirming}
        onConfirm={() => void confirmExport()}
      />
    </Stack>
  );
}
