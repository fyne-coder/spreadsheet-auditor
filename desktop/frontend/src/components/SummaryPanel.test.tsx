import { MantineProvider } from "@mantine/core";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { makeIssue, makeReport } from "../test/fixtures";
import { SummaryPanel } from "./SummaryPanel";

describe("SummaryPanel", () => {
  it("keeps rollup chips collapsed until breakdown is expanded", async () => {
    const user = userEvent.setup();
    const report = makeReport([
      makeIssue({
        RuleID: "RULE_A",
        Message: "first",
        Evidence: { Sheet: "Inputs", Cell: "A1", Formula: "" },
      }),
      makeIssue({
        RuleID: "RULE_B",
        Message: "second",
        Evidence: { Sheet: "Inputs", Cell: "A2", Formula: "" },
      }),
      makeIssue({
        RuleID: "RULE_C",
        Message: "third",
        Evidence: { Sheet: "Calc", Cell: "B1", Formula: "" },
      }),
    ]);

    render(
      <MantineProvider>
        <SummaryPanel report={report} />
      </MantineProvider>,
    );

    expect(screen.getByTestId("summary-breakdown-toggle")).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByTestId("summary-breakdown")).not.toBeInTheDocument();
    expect(screen.queryByText("Inputs: 2")).not.toBeInTheDocument();

    await user.click(screen.getByTestId("summary-breakdown-toggle"));

    expect(screen.getByTestId("summary-breakdown-toggle")).toHaveAttribute("aria-expanded", "true");
    expect(screen.getByTestId("summary-breakdown")).toBeInTheDocument();
    expect(screen.getByText("Sheet issues")).toBeInTheDocument();
    expect(screen.getByText("Inputs: 2")).toBeInTheDocument();
    expect(screen.getByText("Calc: 1")).toBeInTheDocument();
  });

  it("shows key workbook stats while breakdown is collapsed", () => {
    const report = makeReport([
      makeIssue({
        RuleID: "RULE_A",
        Message: "first",
        Severity: "high",
        Evidence: { Sheet: "Inputs", Cell: "A1", Formula: "" },
      }),
    ]);

    render(
      <MantineProvider>
        <SummaryPanel report={report} />
      </MantineProvider>,
    );

    const panel = screen.getByTestId("summary-panel");
    expect(within(panel).getByText("Issues")).toBeInTheDocument();
    expect(within(panel).getByText("High")).toBeInTheDocument();
    expect(within(panel).getAllByText("1").length).toBeGreaterThanOrEqual(1);
  });
});
