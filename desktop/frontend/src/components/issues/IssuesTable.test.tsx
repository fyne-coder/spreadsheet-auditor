import { MantineProvider } from "@mantine/core";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it } from "vitest";
import { makeIssue, makeReport } from "../../test/fixtures";
import { IssuesTable } from "./IssuesTable";

function renderTable(issueCount = 5) {
  const issues = [
    makeIssue({
      RuleID: "RULE_A",
      Message: "alpha message",
      Severity: "high",
      Priority: "high",
      Category: "structure",
      Evidence: { Sheet: "Inputs", Cell: "B2", Formula: "=SUM(A1:A10)" },
    }),
    makeIssue({
      RuleID: "RULE_B",
      Message: "beta message",
      Severity: "medium",
      Priority: "medium",
      Category: "formula",
      Evidence: { Sheet: "Calc", Cell: "C3", Formula: "" },
    }),
    makeIssue({
      RuleID: "RULE_C",
      Message: "gamma hidden",
      Severity: "low",
      Priority: "low",
      Category: "reference",
      Evidence: { Sheet: "Outputs", Cell: "D4", Formula: "" },
    }),
    makeIssue({
      RuleID: "RULE_D",
      Message: "delta note",
      Severity: "high",
      Priority: "high",
      Category: "formula",
      Evidence: { Sheet: "Inputs", Cell: "E5", Formula: "" },
    }),
    makeIssue({
      RuleID: "RULE_E",
      Message: "epsilon tail",
      Severity: "medium",
      Priority: "medium",
      Category: "structure",
      Evidence: { Sheet: "Calc", Cell: "F6", Formula: "" },
    }),
  ].slice(0, issueCount);

  const report = makeReport(issues);
  return render(
    <MantineProvider>
      <IssuesTable issues={report.Issues ?? []} />
    </MantineProvider>,
  );
}

describe("IssuesTable", () => {
  it("sorts by priority when header is clicked", async () => {
    const user = userEvent.setup();
    renderTable();
    const priorityHeader = screen.getByRole("columnheader", { name: /priority/i });
    await user.click(priorityHeader);
    const rows = screen.getAllByTestId(/^issue-row-/);
    expect(rows.length).toBeGreaterThan(0);
    expect(within(rows[0]).getAllByText("high").length).toBeGreaterThan(0);
  });

  it("filters rows with global search", async () => {
    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByTestId("global-search"), "gamma");
    expect(screen.getByText("gamma hidden")).toBeInTheDocument();
    expect(screen.queryByText("alpha message")).not.toBeInTheDocument();
    expect(screen.getByTestId("issue-counts")).toHaveTextContent("1 filtered");
  });

  it("paginates rows", async () => {
    const user = userEvent.setup();
    const many = Array.from({ length: 30 }, (_, i) =>
      makeIssue({
        RuleID: `RULE_${i}`,
        Message: `issue ${i}`,
        Severity: "medium",
        Category: "formula",
        Evidence: { Sheet: "S1", Cell: `A${i + 1}`, Formula: "" },
      }),
    );
    const report = makeReport(many);
    render(
      <MantineProvider>
        <IssuesTable issues={report.Issues ?? []} />
      </MantineProvider>,
    );
    expect(screen.getAllByTestId(/^issue-row-/).length).toBe(25);
    await user.click(screen.getByTestId("next-page"));
    expect(screen.getAllByTestId(/^issue-row-/).length).toBe(5);
  });

  it("opens row detail drawer", async () => {
    const user = userEvent.setup();
    renderTable(1);
    await user.click(screen.getByTestId(/^issue-row-/));
    expect(screen.getByText("Issue")).toBeInTheDocument();
    expect(screen.getByText(/Why this matters/i)).toBeInTheDocument();
    expect(screen.getByText(/Suggested fix/i)).toBeInTheDocument();
    expect(screen.queryByText(/Copy issue key/i)).not.toBeInTheDocument();
    await user.click(screen.getByTestId("issue-advanced-toggle"));
    expect(screen.getByText(/Copy issue key/i)).toBeInTheDocument();
  });

  it("opens row detail drawer from an explicit Details button", async () => {
    const user = userEvent.setup();
    renderTable(1);

    await user.click(screen.getByRole("button", { name: /open details for rule_a/i }));

    expect(screen.getByText("Issue")).toBeInTheDocument();
    expect(screen.getByText(/Why this matters/i)).toBeInTheDocument();
  });

  it("accepts controlled column filters from overview routing", () => {
    const issues = [
      makeIssue({
        RuleID: "RULE_A",
        Message: "alpha message",
        Severity: "high",
        Category: "structure",
      }),
      makeIssue({
        RuleID: "RULE_B",
        Message: "beta message",
        Severity: "medium",
        Category: "formula",
      }),
    ];
    const report = makeReport(issues);

    render(
      <MantineProvider>
        <IssuesTable
          issues={report.Issues ?? []}
          columnFilters={[{ id: "Severity", value: ["high"] }]}
        />
      </MantineProvider>,
    );

    expect(screen.getByText("alpha message")).toBeInTheDocument();
    expect(screen.queryByText("beta message")).not.toBeInTheDocument();
    expect(screen.getByTestId("issue-counts")).toHaveTextContent("1 filtered");
  });

  it("toggles column visibility", async () => {
    const user = userEvent.setup();
    renderTable(1);
    expect(screen.getByRole("columnheader", { name: /formula/i })).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /columns/i }));
    const formulaToggle = await screen.findByRole("menuitem", { name: /formula/i });
    await user.click(within(formulaToggle).getByRole("checkbox"));
    expect(screen.queryByRole("columnheader", { name: /formula/i })).not.toBeInTheDocument();
  });

  it("shows empty filtered state", async () => {
    const user = userEvent.setup();
    renderTable();
    await user.type(screen.getByTestId("global-search"), "no-such-issue-xyz");
    expect(screen.getByTestId("empty-filtered-state")).toBeInTheDocument();
  });

  it("shows zero issues state", () => {
    render(
      <MantineProvider>
        <IssuesTable issues={[]} />
      </MantineProvider>,
    );
    expect(screen.getByTestId("zero-issues-state")).toBeInTheDocument();
  });

  it("uses a viewport-sized scroll region instead of a 420px cap", () => {
    renderTable();
    const scroll = screen.getByTestId("issues-table-scroll");
    expect(scroll.className).toMatch(/tableRegion/);
    expect(scroll).not.toHaveAttribute("mah", "420");
  });

  it("renders long formula and remediation values in truncated one-line cells", () => {
    const longFormula = `=SUM(${ "A".repeat(120) }:Z999)`;
    const longRemediation = `Review ${"and verify ".repeat(20)}the source range.`;
    const issues = [
      makeIssue({
        RuleID: "RULE_LONG",
        Message: "short message",
        Remediation: longRemediation,
        Evidence: { Sheet: "Inputs", Cell: "B2", Formula: longFormula },
      }),
    ];
    render(
      <MantineProvider>
        <IssuesTable issues={issues} />
      </MantineProvider>,
    );

    const truncatedCells = screen
      .getAllByTestId("truncated-cell")
      .filter((cell) => cell.getAttribute("data-truncate") === "true");
    expect(truncatedCells.length).toBeGreaterThanOrEqual(2);
    for (const cell of truncatedCells) {
      expect(cell.className).toMatch(/mantine-Text-root/);
    }
    expect(screen.getByText(longFormula)).toBeInTheDocument();
    expect(screen.getByText(longRemediation)).toBeInTheDocument();
  });

  it("keeps compact filter controls in a single toolbar row", () => {
    renderTable();
    const toolbar = screen.getByTestId("issues-table-toolbar");
    expect(toolbar.className).toMatch(/toolbar/);
    expect(within(toolbar).getByTestId("global-search")).toBeInTheDocument();
    expect(within(toolbar).getByTestId("severity-filter")).toBeInTheDocument();
    expect(within(toolbar).getByRole("button", { name: /columns/i })).toBeInTheDocument();
  });

  it("uses filter triggers and an applied-filter strip instead of inline selected chips", async () => {
    const user = userEvent.setup();
    renderTable();

    const categoryFilter = screen.getByTestId("category-filter");
    expect(categoryFilter).toHaveTextContent("Category");
    await user.click(categoryFilter);
    await user.click(await screen.findByRole("checkbox", { name: "formula" }));

    expect(categoryFilter).toHaveTextContent("Category · 1");
    const appliedFilters = screen.getByTestId("applied-filters");
    expect(appliedFilters).toHaveTextContent("Category: formula");
    expect(within(appliedFilters).getByTestId("clear-all-filters")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /FORMULA_PATTERN_ANOMALY/i })).not.toBeInTheDocument();
  });

  it("removes individual applied filters from the filter strip", async () => {
    const user = userEvent.setup();
    renderTable();

    await user.click(screen.getByTestId("category-filter"));
    await user.click(await screen.findByRole("checkbox", { name: "formula" }));
    expect(screen.getByTestId("applied-filters")).toHaveTextContent("Category: formula");

    await user.click(screen.getByRole("button", { name: "Remove filter Category: formula" }));
    expect(screen.queryByTestId("applied-filters")).not.toBeInTheDocument();
    expect(screen.getByTestId("category-filter")).toHaveTextContent("Category");
  });

  it("clears facet filters without clearing global search", async () => {
    const user = userEvent.setup();
    renderTable();

    await user.type(screen.getByTestId("global-search"), "beta");
    await user.click(screen.getByTestId("category-filter"));
    await user.click(await screen.findByRole("checkbox", { name: "formula" }));
    expect(screen.getByText("beta message")).toBeInTheDocument();

    await user.click(screen.getByTestId("clear-all-filters"));
    expect(screen.queryByTestId("applied-filters")).not.toBeInTheDocument();
    expect(screen.getByTestId("global-search")).toHaveValue("beta");
    expect(screen.getByTestId("issue-counts")).toHaveTextContent("1 filtered");
  });

  it("opens and closes facet popovers from the keyboard", async () => {
    const user = userEvent.setup();
    renderTable();

    const ruleFilter = screen.getByTestId("rule-filter");
    ruleFilter.focus();
    await user.keyboard("{Enter}");
    const ruleOption = await screen.findByRole("checkbox", { name: "RULE_A" });
    expect(ruleOption).toBeInTheDocument();

    ruleOption.focus();
    await user.keyboard("{Escape}");
    await waitFor(() => {
      expect(screen.queryByRole("checkbox", { name: "RULE_A" })).not.toBeInTheDocument();
      expect(ruleFilter).toHaveFocus();
    });
  });
});
