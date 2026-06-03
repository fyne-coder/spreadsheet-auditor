import { MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";
import { issueKey } from "./lib/issueKey";
import { makeIssue, makeReport } from "./test/fixtures";
import * as AuditService from "../wailsjs/go/main/AuditService";

vi.mock("../wailsjs/go/main/AuditService", () => ({
  PickWorkbook: vi.fn(),
  ScanWorkbook: vi.fn(),
  PickExportSavePath: vi.fn(),
  SaveExport: vi.fn(),
}));

const mockScanWorkbook = vi.mocked(AuditService.ScanWorkbook);
const mockPickExportSavePath = vi.mocked(AuditService.PickExportSavePath);
const mockSaveExport = vi.mocked(AuditService.SaveExport);

const FIXED_EXPORTED_AT = "2026-06-02T12:00:00.000Z";

function renderApp() {
  return render(
    <MantineProvider>
      <Notifications />
      <App />
    </MantineProvider>,
  );
}

async function scanFixtureIssues() {
  const user = userEvent.setup();
  const issues = [
    makeIssue({
      RuleID: "RULE_A",
      Message: "alpha",
      Evidence: { Sheet: "Inputs", Cell: "B2", Formula: "" },
    }),
    makeIssue({
      RuleID: "RULE_B",
      Message: "beta",
      Evidence: { Sheet: "Calc", Cell: "C3", Formula: "" },
    }),
  ];
  const report = makeReport(issues);
  mockScanWorkbook.mockResolvedValue(report);

  renderApp();
  await user.type(screen.getByTestId("workbook-path"), "/tmp/sample.xlsx");
  await user.click(screen.getByTestId("scan-btn"));
  await screen.findByText(/2 issues found/i);
  return { user, issues, report };
}

describe("App export flow", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(Date.prototype, "toISOString").mockReturnValue(FIXED_EXPORTED_AT);
    mockPickExportSavePath.mockResolvedValue("/tmp/out/review-pack.html");
    mockSaveExport.mockResolvedValue(undefined);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("opens export modal from scanned state", async () => {
    const { user } = await scanFixtureIssues();
    await user.click(screen.getByTestId("export-btn"));
    expect(screen.getByTestId("export-modal")).toBeInTheDocument();
    expect(screen.getByTestId("export-filename")).toHaveTextContent("review-pack.html");
  });

  it("switches default filename when format changes", async () => {
    const { user } = await scanFixtureIssues();
    await user.click(screen.getByTestId("export-btn"));
    expect(screen.getByTestId("export-filename")).toHaveTextContent("review-pack.html");

    await user.click(screen.getByText("CSV"));
    expect(screen.getByTestId("export-filename")).toHaveTextContent("review-pack.csv");

    await user.click(screen.getByTestId("export-confirm"));
    await waitFor(() => {
      expect(mockPickExportSavePath).toHaveBeenCalledWith("csv", "review-pack.csv");
    });
  });

  it("exports all issues with empty issueIDs", async () => {
    const { user } = await scanFixtureIssues();
    await user.click(screen.getByTestId("export-btn"));
    await user.click(screen.getByTestId("export-confirm"));

    await waitFor(() => {
      expect(mockSaveExport).toHaveBeenCalledWith(
        "/tmp/sample.xlsx",
        "/tmp/out/review-pack.html",
        "html",
        FIXED_EXPORTED_AT,
        [],
      );
    });
    expect(screen.getByTestId("status")).toHaveTextContent(FIXED_EXPORTED_AT);
  });

  it("exports selected issues with issue IDs", async () => {
    const { user, issues } = await scanFixtureIssues();
    const firstId = issueKey(issues[0]);
    const issueSelectors = screen.getAllByLabelText("Select issue");
    await user.click(issueSelectors[0]);
    expect(screen.getByTestId("issue-counts")).toHaveTextContent("1 selected");

    await user.click(screen.getByTestId("export-btn"));
    await waitFor(() => {
      expect(screen.getByText("Selected issues (1)")).toBeInTheDocument();
    });
    await user.click(screen.getByText("Selected issues (1)"));
    await user.click(screen.getByTestId("export-confirm"));

    await waitFor(() => {
      expect(mockSaveExport).toHaveBeenCalledWith(
        "/tmp/sample.xlsx",
        "/tmp/out/review-pack.html",
        "html",
        FIXED_EXPORTED_AT,
        [firstId],
      );
    });
  });

  it("shows exported-at in modal and status on success", async () => {
    let resolveSave: () => void = () => undefined;
    mockSaveExport.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          resolveSave = resolve;
        }),
    );

    const { user } = await scanFixtureIssues();
    await user.click(screen.getByTestId("export-btn"));
    await user.click(screen.getByTestId("export-confirm"));

    await waitFor(() => {
      expect(screen.getByTestId("export-exported-at")).toHaveTextContent(FIXED_EXPORTED_AT);
    });

    resolveSave();
    await waitFor(() => {
      expect(screen.getByTestId("status")).toHaveTextContent(
        `Exported at ${FIXED_EXPORTED_AT}`,
      );
    });
    expect(mockSaveExport).toHaveBeenCalledWith(
      "/tmp/sample.xlsx",
      "/tmp/out/review-pack.html",
      "html",
      FIXED_EXPORTED_AT,
      [],
    );
  });

  it("handles cancelled save dialog without failing", async () => {
    const { user } = await scanFixtureIssues();
    mockPickExportSavePath.mockResolvedValue("");
    await user.click(screen.getByTestId("export-btn"));
    await user.click(screen.getByTestId("export-confirm"));

    await waitFor(() => {
      expect(screen.getByTestId("status")).toHaveTextContent(/cancelled/i);
    });
    expect(mockSaveExport).not.toHaveBeenCalled();
    expect(screen.getByTestId("export-modal")).toBeInTheDocument();
  });

  it("shows backend export failures", async () => {
    const { user } = await scanFixtureIssues();
    mockSaveExport.mockRejectedValue(new Error("permission denied"));

    await user.click(screen.getByTestId("export-btn"));
    await user.click(screen.getByTestId("export-confirm"));

    await waitFor(() => {
      expect(screen.getByTestId("status")).toHaveTextContent("permission denied");
    });
    expect(screen.getByTestId("export-modal")).toBeInTheDocument();
  });

  it("shows a clear message when the desktop bridge is unavailable during scan", async () => {
    const user = userEvent.setup();
    mockScanWorkbook.mockRejectedValue(
      new TypeError("Cannot read properties of undefined (reading 'main')"),
    );

    renderApp();
    await user.type(screen.getByTestId("workbook-path"), "/tmp/sample.xlsx");
    await user.click(screen.getByTestId("scan-btn"));

    await waitFor(() => {
      expect(screen.getByTestId("status")).toHaveTextContent(
        "Desktop bridge unavailable",
      );
    });
    expect(screen.getByTestId("scan-error")).toHaveTextContent(
      "Open the Wails desktop app",
    );
  });
});
