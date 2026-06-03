from __future__ import annotations

from collections import Counter
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


@dataclass(frozen=True)
class RuleDefinition:
    rule_id: str
    title: str
    severity: str
    category: str
    rationale: str
    remediation: str


RULES: dict[str, RuleDefinition] = {
    "BROKEN_REF_VALUE": RuleDefinition(
        rule_id="BROKEN_REF_VALUE",
        title="Broken reference value",
        severity="high",
        category="formula_integrity",
        rationale=(
            "A cell displays #REF!, which usually means a deleted row, column, sheet, "
            "or workbook link left a dangling reference."
        ),
        remediation=(
            "Open the cell and trace the formula to the missing reference. Restore the "
            "source range or replace the reference with a valid sheet/range."
        ),
    ),
    "BROKEN_REF_FORMULA": RuleDefinition(
        rule_id="BROKEN_REF_FORMULA",
        title="Broken reference in formula",
        severity="high",
        category="formula_integrity",
        rationale=(
            "The stored formula text still contains #REF!, so the workbook may calculate "
            "incorrectly even when neighboring cells look fine."
        ),
        remediation=(
            "Repair or remove the broken reference inside the formula, then recalculate "
            "downstream dependents after the source range is restored."
        ),
    ),
    "EXTERNAL_WORKBOOK_REFERENCE": RuleDefinition(
        rule_id="EXTERNAL_WORKBOOK_REFERENCE",
        title="External workbook reference",
        severity="medium",
        category="lineage",
        rationale=(
            "The formula points at another workbook file, which makes results depend on "
            "external file paths, versions, and availability."
        ),
        remediation=(
            "Document the external dependency, confirm the linked file path, or paste "
            "values into a controlled input range when the model must be self-contained."
        ),
    ),
    "FORMULA_PARSE_ERROR": RuleDefinition(
        rule_id="FORMULA_PARSE_ERROR",
        title="Formula parse error",
        severity="medium",
        category="formula_integrity",
        rationale=(
            "The static tokenizer could not read the formula grammar, so additional "
            "lint rules may be skipped for this cell."
        ),
        remediation=(
            "Open the formula in Excel to confirm it is valid, then simplify unusual "
            "syntax or named-range constructs and rerun the audit."
        ),
    ),
    "HARDCODED_NUMERIC_CONSTANT": RuleDefinition(
        rule_id="HARDCODED_NUMERIC_CONSTANT",
        title="Hardcoded numeric constant",
        severity="medium",
        category="formula_integrity",
        rationale=(
            "Numeric literals embedded in formulas are harder to audit, reuse, and "
            "update than values stored in labeled input cells."
        ),
        remediation=(
            "Move constants into clearly labeled input cells or named ranges and "
            "reference those cells from the formula."
        ),
    ),
    "VOLATILE_FUNCTION": RuleDefinition(
        rule_id="VOLATILE_FUNCTION",
        title="Volatile function",
        severity="medium",
        category="performance",
        rationale=(
            "Volatile functions recalculate on every workbook change, which can slow "
            "large models and make timing-dependent results harder to reason about."
        ),
        remediation=(
            "Replace volatile functions where possible, isolate them on a dedicated "
            "calculation sheet, or document why live recalculation is required."
        ),
    ),
    "WHOLE_COLUMN_RANGE": RuleDefinition(
        rule_id="WHOLE_COLUMN_RANGE",
        title="Whole-column range reference",
        severity="medium",
        category="performance",
        rationale=(
            "Entire-column references expand to very large ranges and can slow "
            "recalculation or hide accidental blanks in large worksheets."
        ),
        remediation=(
            "Restrict the range to the populated rows/columns, or convert the logic "
            "to a structured table with explicit bounds."
        ),
    ),
    "FORMULA_PATTERN_ANOMALY": RuleDefinition(
        rule_id="FORMULA_PATTERN_ANOMALY",
        title="Formula pattern anomaly",
        severity="medium",
        category="formula_integrity",
        rationale=(
            "A formula in a copied row/column run does not match the normalized pattern "
            "used by its immediate neighbors, which often signals a paste error or "
            "one-off adjustment."
        ),
        remediation=(
            "Compare the outlier cell with its neighboring formulas, confirm whether the "
            "difference is intentional, and align the formula with the copied pattern "
            "when it is not."
        ),
    ),
}


@dataclass(frozen=True)
class IssueEvidence:
    sheet: str
    cell: str
    formula: str | None = None

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "sheet": self.sheet,
            "cell": self.cell,
        }
        if self.formula is not None:
            payload["formula"] = self.formula
        return payload


@dataclass(frozen=True)
class ImpactFactor:
    code: str
    explanation: str

    def to_dict(self) -> dict[str, str]:
        return {
            "code": self.code,
            "explanation": self.explanation,
        }


@dataclass(frozen=True)
class Issue:
    rule_id: str
    title: str
    severity: str
    category: str
    rationale: str
    remediation: str
    message: str
    evidence: IssueEvidence
    details: dict[str, Any] = field(default_factory=dict)
    priority: str = ""
    impact_factors: tuple[ImpactFactor, ...] = ()

    def to_dict(self) -> dict[str, Any]:
        payload: dict[str, Any] = {
            "rule_id": self.rule_id,
            "title": self.title,
            "severity": self.severity,
            "category": self.category,
            "rationale": self.rationale,
            "remediation": self.remediation,
            "message": self.message,
            "evidence": self.evidence.to_dict(),
        }
        if self.priority:
            payload["priority"] = self.priority
        if self.impact_factors:
            payload["impact_factors"] = [
                factor.to_dict() for factor in self.impact_factors
            ]
        if self.details:
            payload["details"] = {key: self.details[key] for key in sorted(self.details)}
        return payload


def build_issue(
    rule_id: str,
    *,
    message: str,
    sheet: str,
    cell: str,
    formula: str | None = None,
    details: dict[str, Any] | None = None,
) -> Issue:
    rule = RULES[rule_id]
    return Issue(
        rule_id=rule.rule_id,
        title=rule.title,
        severity=rule.severity,
        category=rule.category,
        priority="",
        impact_factors=(),
        rationale=rule.rationale,
        remediation=rule.remediation,
        message=message,
        evidence=IssueEvidence(sheet=sheet, cell=cell, formula=formula),
        details=details or {},
    )


@dataclass(frozen=True)
class SheetSummary:
    name: str
    state: str
    used_range: str
    formula_cells: int

    def to_dict(self) -> dict[str, Any]:
        return {
            "name": self.name,
            "state": self.state,
            "used_range": self.used_range,
            "formula_cells": self.formula_cells,
        }


@dataclass(frozen=True)
class AuditReport:
    workbook_path: Path
    supported_format: str
    sheets: list[SheetSummary]
    issues: list[Issue]

    def to_dict(self) -> dict[str, Any]:
        severity_counts = Counter(issue.severity for issue in self.issues)
        category_counts = Counter(issue.category for issue in self.issues)
        return {
            "workbook_path": str(self.workbook_path),
            "supported_format": self.supported_format,
            "summary": {
                "sheet_count": len(self.sheets),
                "formula_cell_count": sum(sheet.formula_cells for sheet in self.sheets),
                "issue_count": len(self.issues),
                "issues_by_severity": {
                    severity: severity_counts[severity]
                    for severity in sorted(severity_counts)
                },
                "issues_by_category": {
                    category: category_counts[category]
                    for category in sorted(category_counts)
                },
            },
            "sheets": [sheet.to_dict() for sheet in self.sheets],
            "issues": [issue.to_dict() for issue in self.issues],
        }
