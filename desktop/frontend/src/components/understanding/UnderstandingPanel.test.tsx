import { MantineProvider } from "@mantine/core";
import { Notifications } from "@mantine/notifications";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { UnderstandingPanel } from "./UnderstandingPanel";
import * as AuditService from "../../../wailsjs/go/main/AuditService";
import { model } from "../../../wailsjs/go/models";
import { copyText } from "../../lib/copyText";
import { makePromptBundle, makeUnderstandingReport } from "../../test/fixtures";

vi.mock("../../../wailsjs/go/main/AuditService", () => ({
  BuildAIHandoff: vi.fn(),
  ValidateUnderstandingReport: vi.fn(),
  PickJSONSavePath: vi.fn(),
  SaveEvidencePacket: vi.fn(),
  SavePromptBundle: vi.fn(),
  SaveUnderstandingReport: vi.fn(),
}));

const mockBuildAIHandoff = vi.mocked(AuditService.BuildAIHandoff);
const mockValidateUnderstandingReport = vi.mocked(AuditService.ValidateUnderstandingReport);
const mockPickJSONSavePath = vi.mocked(AuditService.PickJSONSavePath);
const mockSaveUnderstandingReport = vi.mocked(AuditService.SaveUnderstandingReport);
const mockCopyText = vi.mocked(copyText);

function renderPanel() {
  return render(
    <MantineProvider>
      <Notifications />
      <UnderstandingPanel workbookPath="/tmp/sample.xlsx" opened onOpenedChange={vi.fn()} />
    </MantineProvider>,
  );
}

function renderCollapsedPanel() {
  return render(
    <MantineProvider>
      <Notifications />
      <UnderstandingPanel workbookPath="/tmp/sample.xlsx" opened={false} onOpenedChange={vi.fn()} />
    </MantineProvider>,
  );
}

vi.mock("../../lib/copyText", () => ({
  copyText: vi.fn().mockResolvedValue(undefined),
}));

function makeAIHandoff(bundle = makePromptBundle()) {
  return new model.AIHandoffPayload({
    audit_hash: bundle.evidence_packet?.audit_hash ?? "abc",
    prompt: bundle.prompt,
    evidence_packet_json: '{"packet":"canonical"}',
    prompt_bundle_json: '{"bundle":"canonical"}',
    bundle,
  });
}

describe("UnderstandingPanel", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockBuildAIHandoff.mockResolvedValue(makeAIHandoff());
    mockPickJSONSavePath.mockResolvedValue("/tmp/verified-ai-analysis.json");
    mockSaveUnderstandingReport.mockResolvedValue(undefined);
  });

  it("builds preview from a single BuildAIHandoff call", async () => {
    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalledTimes(1));
    expect(screen.getByTestId("understanding-preview")).toHaveTextContent(
      "prompt with exclusions",
    );
  });

  it("stays collapsed until the user starts the AI handoff", () => {
    renderCollapsedPanel();
    expect(screen.getByText("Use with your AI assistant")).toBeInTheDocument();
    expect(screen.getByTestId("understanding-toggle")).toHaveTextContent("Start AI handoff");
    expect(mockBuildAIHandoff).not.toHaveBeenCalled();
    expect(screen.queryByTestId("understanding-preview")).not.toBeInTheDocument();
  });

  it("rebuilds preview when sheet exclusions change", async () => {
    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    fireEvent.change(screen.getByTestId("exclude-sheets"), {
      target: { value: "HiddenInputs" },
    });
    await waitFor(() =>
      expect(mockBuildAIHandoff).toHaveBeenLastCalledWith(
        "/tmp/sample.xlsx",
        expect.objectContaining({ exclude_sheets: ["HiddenInputs"] }),
      ),
    );
    expect(mockBuildAIHandoff).toHaveBeenCalledTimes(2);
  });

  it("includes the user objective in prompt bundle options", async () => {
    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    fireEvent.change(screen.getByTestId("user-objective"), {
      target: { value: "Understand the workbook state before budget planning." },
    });

    await waitFor(() =>
      expect(mockBuildAIHandoff).toHaveBeenLastCalledWith(
        "/tmp/sample.xlsx",
        expect.objectContaining({
          user_objective: "Understand the workbook state before budget planning.",
        }),
      ),
    );
  });

  it("lets the user raise the packet cap when a workbook exceeds the default", async () => {
    const user = userEvent.setup();
    mockBuildAIHandoff
      .mockRejectedValueOnce(
        new Error("evidence packet exceeds max_packet_bytes cap (757093 > 524288)"),
      )
      .mockResolvedValueOnce(makeAIHandoff());

    renderPanel();

    expect(await screen.findByTestId("understanding-error")).toHaveTextContent(
      "Evidence package is over the current size limit",
    );
    expect(screen.getByTestId("understanding-error")).toHaveTextContent("739 KB");
    expect(screen.getByTestId("understanding-error")).toHaveTextContent("512 KB");

    await user.click(screen.getByTestId("increase-packet-limit"));

    await waitFor(() =>
      expect(mockBuildAIHandoff).toHaveBeenLastCalledWith(
        "/tmp/sample.xlsx",
        expect.objectContaining({ max_packet_bytes: 30 * 1024 * 1024 }),
      ),
    );
    expect(await screen.findByTestId("understanding-preview")).toHaveTextContent(
      "prompt with exclusions",
    );
  });

  it("explains provider upload limits without implying spreadsheets can use the full ChatGPT cap", async () => {
    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    expect(screen.getByTestId("packet-size-guidance")).toHaveTextContent(
      "Prefer the smaller evidence packet",
    );
    expect(screen.getByTestId("packet-size-guidance")).toHaveTextContent("512 MB");
    expect(screen.getByTestId("packet-size-guidance")).toHaveTextContent("50 MB");
    expect(screen.getByTestId("packet-size-guidance")).toHaveTextContent(
      "token or context",
    );
  });

  it("supports the ChatGPT file-size cap when selected intentionally", async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    await user.click(screen.getByText("ChatGPT file cap"));

    await waitFor(() =>
      expect(mockBuildAIHandoff).toHaveBeenLastCalledWith(
        "/tmp/sample.xlsx",
        expect.objectContaining({ max_packet_bytes: 512 * 1024 * 1024 }),
      ),
    );
  });

  it("copies backend canonical JSON strings for packet and bundle", async () => {
    const user = userEvent.setup();
    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    await user.click(screen.getByTestId("copy-evidence-packet"));
    expect(mockCopyText).toHaveBeenCalledWith(
      "Evidence packet JSON",
      '{"packet":"canonical"}',
    );

    await user.click(screen.getByTestId("copy-prompt-bundle"));
    expect(mockCopyText).toHaveBeenCalledWith(
      "Prompt bundle JSON",
      '{"bundle":"canonical"}',
    );
  });

  it("renders valid pasted understanding separately from issues", async () => {
    const user = userEvent.setup();
    const report = makeUnderstandingReport("issue-1");
    mockValidateUnderstandingReport.mockResolvedValue(
      new model.UnderstandingValidationResult({
        citations_resolved: true,
        valid: true,
        report,
      }),
    );

    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    fireEvent.change(screen.getByTestId("understanding-paste-input"), {
      target: { value: '{"workbook_purpose":[]}' },
    });
    expect(screen.getByTestId("understanding-validate")).toHaveTextContent(
      "Verify cited evidence",
    );
    await user.click(screen.getByTestId("understanding-validate"));

    await screen.findByTestId("understanding-report-view");
    expect(screen.getByTestId("understanding-accepted")).toHaveTextContent(
      "Cited evidence verified",
    );
    expect(screen.getByTestId("understanding-accepted")).toHaveTextContent(
      "Review the AI's claims before relying on them",
    );
    expect(screen.getByText("Purpose")).toBeInTheDocument();
    expect(screen.queryByTestId("understanding-rejected")).not.toBeInTheDocument();
  });

  it("saves verified JSON and copies Markdown after cited evidence resolves", async () => {
    const user = userEvent.setup();
    const report = makeUnderstandingReport("issue-1");
    mockValidateUnderstandingReport.mockResolvedValue(
      new model.UnderstandingValidationResult({
        citations_resolved: true,
        valid: true,
        report,
      }),
    );

    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());

    fireEvent.change(screen.getByTestId("understanding-paste-input"), {
      target: { value: '{"workbook_purpose":[]}' },
    });
    await user.click(screen.getByTestId("understanding-validate"));
    await screen.findByTestId("understanding-report-view");

    await user.click(screen.getByTestId("save-understanding-report"));
    expect(mockSaveUnderstandingReport).toHaveBeenCalledWith(
      "/tmp/sample.xlsx",
      '{"workbook_purpose":[]}',
      "/tmp/verified-ai-analysis.json",
      expect.any(Object),
    );

    await user.click(screen.getByTestId("copy-understanding-report"));
    expect(mockCopyText).toHaveBeenCalledWith(
      "Verified AI analysis Markdown",
      expect.stringContaining("# Verified AI Analysis"),
    );
    expect(mockCopyText).toHaveBeenCalledWith(
      "Verified AI analysis Markdown",
      expect.stringContaining("Cited evidence check passed"),
    );
    expect(mockCopyText).toHaveBeenCalledWith(
      "Verified AI analysis Markdown",
      expect.stringContaining("`issue-1`"),
    );
  });

  it("rejects fabricated citations", async () => {
    const user = userEvent.setup();
    mockValidateUnderstandingReport.mockResolvedValue(
      new model.UnderstandingValidationResult({
        citations_resolved: false,
        valid: false,
        rejects: [
          new model.CitationReject({
            citation: "fake_issue",
            field: "workbook_purpose.citations",
            index: 0,
            reason: "citation is not present in evidence packet citation map",
          }),
        ],
      }),
    );

    renderPanel();
    await waitFor(() => expect(mockBuildAIHandoff).toHaveBeenCalled());
    fireEvent.change(screen.getByTestId("understanding-paste-input"), {
      target: { value: "{}" },
    });
    await user.click(screen.getByTestId("understanding-validate"));

    expect(await screen.findByTestId("understanding-rejected")).toBeInTheDocument();
    expect(screen.getByText(/fake_issue/)).toBeInTheDocument();
  });
});
