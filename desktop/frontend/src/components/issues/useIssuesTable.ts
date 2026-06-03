import {
  getCoreRowModel,
  getFacetedRowModel,
  getFacetedUniqueValues,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  useReactTable,
  type ColumnFiltersState,
  type OnChangeFn,
  type RowSelectionState,
  type SortingState,
  type VisibilityState,
} from "@tanstack/react-table";
import { useMemo, useState } from "react";
import type { model } from "../../../wailsjs/go/models";
import { matchesGlobalFilter } from "../../lib/globalFilter";
import { issueColumns, toIssueRows } from "./issueColumns";

export type IssuesTableState = ReturnType<typeof useIssuesTable>;

type UseIssuesTableOptions = {
  rowSelection?: RowSelectionState;
  onRowSelectionChange?: OnChangeFn<RowSelectionState>;
};

export function useIssuesTable(issues: model.Issue[], options: UseIssuesTableOptions = {}) {
  const data = useMemo(() => toIssueRows(issues), [issues]);
  const [sorting, setSorting] = useState<SortingState>([
    { id: "Severity", desc: true },
  ]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = useState("");
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const [internalRowSelection, setInternalRowSelection] = useState<RowSelectionState>({});
  const rowSelection = options.rowSelection ?? internalRowSelection;
  const onRowSelectionChange = options.onRowSelectionChange ?? setInternalRowSelection;
  const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 25 });

  const table = useReactTable({
    data,
    columns: issueColumns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
      columnVisibility,
      rowSelection,
      pagination,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getFacetedRowModel: getFacetedRowModel(),
    getFacetedUniqueValues: getFacetedUniqueValues(),
    getPaginationRowModel: getPaginationRowModel(),
    globalFilterFn: (row, _columnId, filterValue) =>
      matchesGlobalFilter(row.original, String(filterValue)),
    enableRowSelection: true,
    getRowId: (row) => row.id,
  });

  const filteredCount = table.getFilteredRowModel().rows.length;
  const selectedCount = Object.keys(rowSelection).length;

  return {
    table,
    globalFilter,
    setGlobalFilter,
    filteredCount,
    selectedCount,
    totalCount: issues.length,
  };
}
