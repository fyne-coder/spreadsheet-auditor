import {
  Badge,
  Box,
  Button,
  Checkbox,
  CloseButton,
  Group,
  Menu,
  Popover,
  ScrollArea,
  Select,
  Stack,
  Table,
  Text,
  TextInput,
} from "@mantine/core";
import { flexRender, type OnChangeFn, type RowSelectionState } from "@tanstack/react-table";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { model } from "../../../wailsjs/go/models";
import { columnLabels } from "./issueColumns";
import classes from "./issuesTable.module.css";
import { IssueDetailDrawer } from "./IssueDetailDrawer";
import { useIssuesTable } from "./useIssuesTable";

type Props = {
  issues: model.Issue[];
  rowSelection?: RowSelectionState;
  onRowSelectionChange?: OnChangeFn<RowSelectionState>;
};

function severityColor(severity: string): string {
  switch (severity) {
    case "high":
      return "red";
    case "medium":
      return "orange";
    case "low":
      return "blue";
    default:
      return "gray";
  }
}

type FacetOption = { value: string; label: string };

type FacetConfig = {
  id: string;
  label: string;
  testId: string;
  width?: number;
};

function facetOptions(
  table: ReturnType<typeof useIssuesTable>["table"],
  columnId: string,
  selectedValues: string[] = [],
): FacetOption[] {
  const column = table.getColumn(columnId);
  if (!column) {
    return [];
  }
  return Array.from(new Set([...Array.from(column.getFacetedUniqueValues().keys()), ...selectedValues]))
    .filter((v) => v !== "")
    .sort()
    .map((value) => ({ value, label: value }));
}

function getFacetValues(table: ReturnType<typeof useIssuesTable>["table"], columnId: string) {
  return (table.getColumn(columnId)?.getFilterValue() as string[]) ?? [];
}

function setFacetValues(
  table: ReturnType<typeof useIssuesTable>["table"],
  columnId: string,
  values: string[],
) {
  table.getColumn(columnId)?.setFilterValue(values.length > 0 ? values : undefined);
}

type FacetFilterProps = {
  config: FacetConfig;
  options: FacetOption[];
  selectedValues: string[];
  onChange: (values: string[]) => void;
  onClear: () => void;
};

function FacetFilter({ config, options, selectedValues, onChange, onClear }: FacetFilterProps) {
  const [search, setSearch] = useState("");
  const [opened, setOpened] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const filteredOptions = useMemo(() => {
    const normalized = search.trim().toLowerCase();
    if (!normalized) {
      return options;
    }
    return options.filter((option) => option.label.toLowerCase().includes(normalized));
  }, [options, search]);
  const active = selectedValues.length > 0;
  const label = active ? `${config.label} · ${selectedValues.length}` : config.label;
  const closePopover = useCallback(() => {
    setOpened(false);
    requestAnimationFrame(() => triggerRef.current?.focus());
  }, []);

  useEffect(() => {
    if (!opened) {
      return undefined;
    }
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        closePopover();
      }
    };
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [closePopover, opened]);

  return (
    <Popover
      width={config.width ?? 240}
      trapFocus
      returnFocus
      shadow="md"
      position="bottom-start"
      opened={opened}
      onChange={setOpened}
    >
      <Popover.Target>
        <Button
          ref={triggerRef}
          variant={active ? "light" : "default"}
          size="compact-xs"
          className={classes.filterTrigger}
          data-testid={config.testId}
          aria-label={`Filter ${config.label}`}
          onClick={() => setOpened((current) => !current)}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              closePopover();
            }
          }}
        >
          {label}
        </Button>
      </Popover.Target>
      <Popover.Dropdown
        aria-label={`Filter ${config.label}`}
        onKeyDown={(event) => {
          if (event.key === "Escape") {
            closePopover();
          }
        }}
      >
        <Stack gap="xs">
          <Group justify="space-between" gap="xs" wrap="nowrap">
            <Text fw={600} size="xs">
              {config.label}
            </Text>
            {active ? (
              <Button variant="subtle" size="compact-xs" onClick={onClear}>
                Clear
              </Button>
            ) : null}
          </Group>
          {options.length > 8 ? (
            <TextInput
              size="xs"
              placeholder={`Search ${config.label.toLowerCase()}`}
              value={search}
              onChange={(event) => setSearch(event.currentTarget.value)}
              aria-label={`Search ${config.label} filters`}
            />
          ) : null}
          <ScrollArea.Autosize mah={280}>
            <Checkbox.Group value={selectedValues} onChange={onChange}>
              <Stack gap={6}>
                {filteredOptions.map((option) => (
                  <Checkbox
                    key={option.value}
                    value={option.value}
                    label={option.label}
                    size="xs"
                    classNames={{ label: classes.filterOptionLabel }}
                  />
                ))}
              </Stack>
            </Checkbox.Group>
          </ScrollArea.Autosize>
        </Stack>
      </Popover.Dropdown>
    </Popover>
  );
}

export function IssuesTable({ issues, rowSelection, onRowSelectionChange }: Props) {
  const { table, globalFilter, setGlobalFilter, filteredCount, selectedCount, totalCount } =
    useIssuesTable(issues, { rowSelection, onRowSelectionChange });
  const [detailIssue, setDetailIssue] = useState<model.Issue | null>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);

  const facetConfigs: FacetConfig[] = useMemo(
    () => [
      { id: "Severity", label: "Severity", testId: "severity-filter" },
      { id: "Category", label: "Category", testId: "category-filter" },
      { id: "RuleID", label: "Rule", testId: "rule-filter", width: 320 },
      { id: "sheet", label: "Sheet", testId: "sheet-filter" },
    ],
    [],
  );
  const activeFilters = facetConfigs.flatMap((config) =>
    getFacetValues(table, config.id).map((value) => ({ ...config, value })),
  );
  const hasActiveFilters = activeFilters.length > 0;

  const hideableColumns = table
    .getAllColumns()
    .filter((col) => col.getCanHide() && col.id !== "select");

  if (totalCount === 0) {
    return (
      <Box p="md" data-testid="zero-issues-state">
        <Text fw={600}>No issues found</Text>
        <Text size="sm" c="dimmed">
          The scan completed successfully but the workbook has no analyzer issues.
        </Text>
      </Box>
    );
  }

  return (
    <Box className={classes.root} data-testid="issues-table-root">
      <Group
        justify="space-between"
        align="center"
        className={classes.toolbar}
        wrap="nowrap"
        gap="xs"
        data-testid="issues-table-toolbar"
      >
        <Text size="xs" data-testid="issue-counts" style={{ whiteSpace: "nowrap", flexShrink: 0 }}>
          {totalCount} total · {filteredCount} filtered · {selectedCount} selected
        </Text>
        <Group gap={6} wrap="nowrap" style={{ flex: 1, justifyContent: "flex-end", minWidth: 0 }}>
          <TextInput
            placeholder="Search issues…"
            value={globalFilter}
            onChange={(e) => setGlobalFilter(e.currentTarget.value)}
            aria-label="Global search"
            data-testid="global-search"
            size="xs"
            className={classes.searchControl}
          />
          {facetConfigs.map((config) => {
            const selectedValues = getFacetValues(table, config.id);
            return (
              <FacetFilter
                key={config.id}
                config={config}
                options={facetOptions(table, config.id, selectedValues)}
                selectedValues={selectedValues}
                onChange={(values) => setFacetValues(table, config.id, values)}
                onClear={() => setFacetValues(table, config.id, [])}
              />
            );
          })}
          <Menu shadow="md" width={200}>
            <Menu.Target>
              <Button variant="default" size="compact-xs">
                Columns
              </Button>
            </Menu.Target>
            <Menu.Dropdown data-testid="column-visibility-menu">
              {hideableColumns.map((column) => (
                <Menu.Item
                  key={column.id}
                  closeMenuOnClick={false}
                  onClick={() => column.toggleVisibility()}
                >
                  <label style={{ display: "flex", gap: 8, cursor: "pointer" }}>
                    <input
                      type="checkbox"
                      checked={column.getIsVisible()}
                      onChange={column.getToggleVisibilityHandler()}
                      onClick={(e) => e.stopPropagation()}
                    />
                    {columnLabels[column.id] ?? column.id}
                  </label>
                </Menu.Item>
              ))}
            </Menu.Dropdown>
          </Menu>
        </Group>
      </Group>

      {hasActiveFilters ? (
        <Group gap={4} wrap="wrap" className={classes.appliedFilters} data-testid="applied-filters">
          <Text size="xs" c="dimmed">
            Filters:
          </Text>
          {activeFilters.map((filter) => (
            <Group
              key={`${filter.id}-${filter.value}`}
              gap={4}
              wrap="nowrap"
              title={`${filter.label}: ${filter.value}`}
              data-testid={`applied-filter-chip-${filter.id}-${filter.value}`}
              className={classes.appliedFilterPill}
            >
              <Text size="xs" className={classes.appliedFilterLabel}>
                {filter.label}: {filter.value}
              </Text>
              <CloseButton
                size="xs"
                aria-label={`Remove filter ${filter.label}: ${filter.value}`}
                onClick={() =>
                  setFacetValues(
                    table,
                    filter.id,
                    getFacetValues(table, filter.id).filter((value) => value !== filter.value),
                  )
                }
              />
            </Group>
          ))}
          <Button
            variant="subtle"
            size="compact-xs"
            onClick={() => table.resetColumnFilters()}
            data-testid="clear-all-filters"
          >
            Clear all
          </Button>
        </Group>
      ) : null}

      {filteredCount === 0 ? (
        <Box p="md" data-testid="empty-filtered-state">
          <Text fw={600}>No issues match the current filters</Text>
          <Text size="sm" c="dimmed">
            Clear search or facet filters to see issues again.
          </Text>
        </Box>
      ) : (
        <>
          <ScrollArea
            className={classes.tableRegion}
            type="auto"
            offsetScrollbars
            data-testid="issues-table-scroll"
          >
            <Table
              striped
              highlightOnHover
              withTableBorder
              stickyHeader
              data-testid="issues-table"
              className={classes.compactTable}
              verticalSpacing="xs"
              horizontalSpacing="xs"
            >
              <Table.Thead>
                {table.getHeaderGroups().map((headerGroup) => (
                  <Table.Tr key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <Table.Th
                        key={header.id}
                        onClick={header.column.getCanSort() ? header.column.getToggleSortingHandler() : undefined}
                        style={{
                          cursor: header.column.getCanSort() ? "pointer" : "default",
                          whiteSpace: "nowrap",
                        }}
                      >
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                        {header.column.getIsSorted() === "asc"
                          ? " ↑"
                          : header.column.getIsSorted() === "desc"
                            ? " ↓"
                            : null}
                      </Table.Th>
                    ))}
                  </Table.Tr>
                ))}
              </Table.Thead>
              <Table.Tbody>
                {table.getRowModel().rows.map((row) => (
                  <Table.Tr
                    key={row.id}
                    onClick={() => {
                      setDetailIssue(row.original);
                      setDrawerOpen(true);
                    }}
                    style={{ cursor: "pointer" }}
                    data-testid={`issue-row-${row.id}`}
                  >
                    {row.getVisibleCells().map((cell) => (
                      <Table.Td key={cell.id}>
                        {cell.column.id === "Severity" ? (
                          <Badge color={severityColor(row.original.Severity)} variant="light" size="xs">
                            {row.original.Severity}
                          </Badge>
                        ) : (
                          flexRender(cell.column.columnDef.cell, cell.getContext())
                        )}
                      </Table.Td>
                    ))}
                  </Table.Tr>
                ))}
              </Table.Tbody>
            </Table>
          </ScrollArea>

          <Group justify="space-between" mt="xs" align="center">
            <Text size="xs" c="dimmed">
              Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount()}
            </Text>
            <Group gap={6}>
              <Select
                size="xs"
                w={72}
                data={["10", "25", "50", "100"]}
                value={String(table.getState().pagination.pageSize)}
                onChange={(value) => table.setPageSize(Number(value ?? 25))}
                aria-label="Page size"
                data-testid="page-size"
              />
              <Button
                size="compact-xs"
                variant="default"
                onClick={() => table.previousPage()}
                disabled={!table.getCanPreviousPage()}
                data-testid="prev-page"
              >
                Previous
              </Button>
              <Button
                size="compact-xs"
                variant="default"
                onClick={() => table.nextPage()}
                disabled={!table.getCanNextPage()}
                data-testid="next-page"
              >
                Next
              </Button>
            </Group>
          </Group>
        </>
      )}

      <IssueDetailDrawer
        issue={detailIssue}
        opened={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      />
    </Box>
  );
}
