from __future__ import annotations

import re
from collections import Counter
from collections.abc import Iterable, Sequence
from dataclasses import dataclass

from openpyxl.formula import Tokenizer
from openpyxl.utils.cell import column_index_from_string, coordinate_from_string, get_column_letter

from spreadsheet_auditor.models import Issue, build_issue

CELL_REFERENCE_RE = re.compile(
    r"(?P<col_abs>\$)?(?P<col>[A-Z]{1,3})(?P<row_abs>\$)?(?P<row>\d+)",
    re.IGNORECASE,
)
MIN_CLUSTER_SIZE = 3


@dataclass(frozen=True)
class FormulaCellRecord:
    coordinate: str
    row: int
    column: int
    formula: str
    pattern: str


def normalize_formula(formula: str, anchor_cell: str) -> str | None:
    """Return a position-relative pattern for comparing copied formulas."""
    try:
        anchor_col, anchor_row = coordinate_from_string(anchor_cell)
        anchor_col_index = column_index_from_string(anchor_col)
        parts: list[str] = []
        for token in Tokenizer(formula).items:
            if token.type == "OPERAND" and token.subtype == "RANGE":
                parts.append(
                    _normalize_range_operand(token.value, anchor_col_index, anchor_row)
                )
            else:
                parts.append(token.value)
        return "".join(parts)
    except Exception:  # pragma: no cover - unusual Excel grammar
        return None


def find_formula_pattern_anomalies(
    sheet: str,
    records: Sequence[FormulaCellRecord],
) -> list[Issue]:
    """Flag a single local outlier inside a copied formula run."""
    issues: list[Issue] = []
    flagged: set[str] = set()

    by_column: dict[int, list[FormulaCellRecord]] = {}
    by_row: dict[int, list[FormulaCellRecord]] = {}
    for record in records:
        by_column.setdefault(record.column, []).append(record)
        by_row.setdefault(record.row, []).append(record)

    for cluster in _column_clusters(by_column):
        issue = _anomaly_issue_for_cluster(sheet, cluster)
        if issue is not None and issue.evidence.cell not in flagged:
            flagged.add(issue.evidence.cell)
            issues.append(issue)

    for cluster in _row_clusters(by_row):
        issue = _anomaly_issue_for_cluster(sheet, cluster)
        if issue is not None and issue.evidence.cell not in flagged:
            flagged.add(issue.evidence.cell)
            issues.append(issue)

    return sorted(issues, key=lambda issue: issue.evidence.cell)


def _normalize_range_operand(value: str, anchor_col: int, anchor_row: int) -> str:
    sheet_prefix = ""
    reference = value
    if "!" in value:
        sheet_prefix, _, reference = value.partition("!")
        sheet_prefix = f"{sheet_prefix}!"

    if ":" in reference:
        start, end = reference.split(":", 1)
        return (
            f"{sheet_prefix}{_normalize_cell_reference(start, anchor_col, anchor_row)}"
            f":{_normalize_cell_reference(end, anchor_col, anchor_row)}"
        )
    return f"{sheet_prefix}{_normalize_cell_reference(reference, anchor_col, anchor_row)}"


def _normalize_cell_reference(reference: str, anchor_col: int, anchor_row: int) -> str:
    match = CELL_REFERENCE_RE.fullmatch(reference)
    if match is None:
        return reference

    column = column_index_from_string(match.group("col"))
    row = int(match.group("row"))
    if match.group("col_abs"):
        column_part = f"${get_column_letter(column)}"
    else:
        column_part = str(column - anchor_col)
    if match.group("row_abs"):
        row_part = f"${row}"
    else:
        row_part = str(row - anchor_row)
    return f"R{row_part}C{column_part}"


def _column_clusters(
    records_by_column: dict[int, list[FormulaCellRecord]],
) -> Iterable[list[FormulaCellRecord]]:
    for column_records in records_by_column.values():
        ordered = sorted(column_records, key=lambda record: record.row)
        yield from _split_consecutive_clusters(ordered, key=lambda record: record.row)


def _row_clusters(
    records_by_row: dict[int, list[FormulaCellRecord]],
) -> Iterable[list[FormulaCellRecord]]:
    for row_records in records_by_row.values():
        ordered = sorted(row_records, key=lambda record: record.column)
        yield from _split_consecutive_clusters(ordered, key=lambda record: record.column)


def _split_consecutive_clusters(
    ordered_records: Sequence[FormulaCellRecord],
    *,
    key,
) -> Iterable[list[FormulaCellRecord]]:
    if not ordered_records:
        return

    cluster = [ordered_records[0]]
    previous = key(ordered_records[0])
    for record in ordered_records[1:]:
        current = key(record)
        if current == previous + 1:
            cluster.append(record)
        else:
            if len(cluster) >= MIN_CLUSTER_SIZE:
                yield cluster
            cluster = [record]
        previous = current
    if len(cluster) >= MIN_CLUSTER_SIZE:
        yield cluster


def _anomaly_issue_for_cluster(
    sheet: str,
    cluster: Sequence[FormulaCellRecord],
) -> Issue | None:
    pattern_counts = Counter(record.pattern for record in cluster)
    if len(pattern_counts) < 2:
        return None

    sorted_patterns = pattern_counts.most_common()
    majority_pattern, majority_count = sorted_patterns[0]
    minority_patterns = sorted_patterns[1:]
    minority_total = sum(count for _, count in minority_patterns)
    if majority_count != len(cluster) - 1 or minority_total != 1:
        return None

    outlier_pattern = minority_patterns[0][0]
    outlier = next(record for record in cluster if record.pattern == outlier_pattern)
    cluster_coordinates = [record.coordinate for record in cluster]
    return build_issue(
        "FORMULA_PATTERN_ANOMALY",
        message=(
            "Formula pattern differs from neighboring copied formulas in the same row/column run."
        ),
        sheet=sheet,
        cell=outlier.coordinate,
        formula=outlier.formula,
        details={
            "cluster_cells": cluster_coordinates,
            "cluster_orientation": (
                "column"
                if len({record.column for record in cluster}) == 1
                else "row"
            ),
            "expected_pattern": majority_pattern,
            "local_pattern": outlier_pattern,
        },
    )
