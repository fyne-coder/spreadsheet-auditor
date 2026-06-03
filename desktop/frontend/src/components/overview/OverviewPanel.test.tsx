import { MantineProvider } from "@mantine/core";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { makeIssue, makeReport } from "../../test/fixtures";
import { OverviewPanel } from "./OverviewPanel";

describe("OverviewPanel", () => {
  it("summarizes scan findings and routes priority buckets to issue filters", async () => {
    const user = userEvent.setup();
    const onApplyFilters = vi.fn();
    const onStartAIHandoff = vi.fn();
    const report = makeReport([
      makeIssue({
        RuleID: "BROKEN_REF_FORMULA",
        Severity: "high",
        Priority: "high",
        Category: "formula_integrity",
        Message: "broken",
      }),
      makeIssue({
        RuleID: "VOLATILE_FUNCTION",
        Severity: "medium",
        Priority: "medium",
        Category: "performance",
        Message: "volatile",
      }),
    ]);

    render(
      <MantineProvider>
        <OverviewPanel
          report={report}
          onApplyFilters={onApplyFilters}
          onStartAIHandoff={onStartAIHandoff}
        />
      </MantineProvider>,
    );

    expect(screen.getByText("Audit overview")).toBeInTheDocument();
    expect(screen.getByTestId("overview-ai-handoff")).toHaveTextContent("AI package");
    expect(screen.getByTestId("overview-panel")).toHaveTextContent(
      "Found 2 issues across 1 sheets",
    );
    expect(screen.getByTestId("overview-panel")).toHaveTextContent(
      "These paths are meant to route review effort.",
    );

    await user.click(screen.getByTestId("overview-bucket-high"));
    expect(onApplyFilters).toHaveBeenCalledWith([{ id: "Priority", value: ["high"] }]);

    await user.click(screen.getByTestId("overview-ai-handoff"));
    expect(onStartAIHandoff).toHaveBeenCalled();
  });
});
