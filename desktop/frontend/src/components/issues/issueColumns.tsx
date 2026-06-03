import { createColumnHelper } from "@tanstack/react-table";
import { model } from "../../../wailsjs/go/models";
import { issueKey } from "../../lib/issueKey";
import { TruncatedCell } from "./TruncatedCell";

export type IssueRow = model.Issue & { id: string };

export function toIssueRows(issues: model.Issue[]): IssueRow[] {
  return issues.map((issue) => {
    const row = new model.Issue(issue) as IssueRow;
    row.id = issueKey(issue);
    return row;
  });
}

const columnHelper = createColumnHelper<IssueRow>();

export const issueColumns = [
  columnHelper.display({
    id: "select",
    header: ({ table }) => (
      <input
        type="checkbox"
        aria-label="Select all rows on page"
        checked={table.getIsAllPageRowsSelected()}
        ref={(el) => {
          if (el) {
            el.indeterminate =
              table.getIsSomePageRowsSelected() && !table.getIsAllPageRowsSelected();
          }
        }}
        onChange={table.getToggleAllPageRowsSelectedHandler()}
      />
    ),
    cell: ({ row }) => (
      <input
        type="checkbox"
        aria-label="Select issue"
        title={row.original.id}
        checked={row.getIsSelected()}
        onChange={row.getToggleSelectedHandler()}
        onClick={(e) => e.stopPropagation()}
      />
    ),
    enableSorting: false,
    enableHiding: false,
  }),
  columnHelper.accessor("Severity", {
    header: "Severity",
    cell: (info) => info.getValue(),
    filterFn: "arrIncludesSome",
  }),
  columnHelper.accessor("Category", {
    header: "Category",
    filterFn: "arrIncludesSome",
  }),
  columnHelper.accessor("RuleID", {
    header: "Rule",
    filterFn: "arrIncludesSome",
  }),
  columnHelper.accessor((row) => row.Evidence?.Sheet ?? "", {
    id: "sheet",
    header: "Sheet",
    filterFn: "arrIncludesSome",
  }),
  columnHelper.accessor((row) => row.Evidence?.Cell ?? "", {
    id: "cell",
    header: "Cell",
  }),
  columnHelper.accessor((row) => row.Evidence?.Formula ?? "", {
    id: "formula",
    header: "Formula",
    cell: (info) => <TruncatedCell value={info.getValue()} maxWidth={200} />,
  }),
  columnHelper.accessor("Message", {
    header: "Message",
    cell: (info) => <TruncatedCell value={info.getValue()} maxWidth={240} />,
  }),
  columnHelper.accessor("Remediation", {
    header: "Remediation",
    cell: (info) => <TruncatedCell value={info.getValue()} maxWidth={240} />,
  }),
];

export const columnLabels: Record<string, string> = {
  select: "Select",
  Severity: "Severity",
  Category: "Category",
  RuleID: "Rule",
  sheet: "Sheet",
  cell: "Cell",
  formula: "Formula",
  Message: "Message",
  Remediation: "Remediation",
};
