from __future__ import annotations

from dataclasses import replace

from spreadsheet_auditor.models import ImpactFactor, Issue, SheetSummary

PRIORITY_CRITICAL = "critical"
PRIORITY_HIGH = "high"
PRIORITY_MEDIUM = "medium"
PRIORITY_LOW = "low"

STRUCTURAL_RULE_IDS = {
    "BROKEN_REF_VALUE",
    "BROKEN_REF_FORMULA",
    "EXCEL_ERROR_VALUE",
    "EXCEL_ERROR_FORMULA",
}

RISKY_VOLATILE_FUNCTIONS = {
    "CELL",
    "INDIRECT",
    "INFO",
    "NOW",
    "OFFSET",
    "RAND",
    "RANDBETWEEN",
}

LOW_RISK_VOLATILE_FUNCTIONS = {"TODAY"}

OUTPUT_LIKE_SHEET_TOKENS = (
    "output",
    "summary",
    "report",
    "dashboard",
    "total",
    "result",
)


def apply_priorities(issues: list[Issue], sheets: list[SheetSummary]) -> list[Issue]:
    states = {sheet.name: sheet.state for sheet in sheets}
    return [
        _assign_issue(issue, states.get(issue.evidence.sheet, "visible"))
        for issue in issues
    ]


def _assign_issue(issue: Issue, sheet_state: str) -> Issue:
    factors = _collect_impact_factors(issue, sheet_state)
    rank = _severity_base_rank(issue.severity)

    if issue.rule_id in STRUCTURAL_RULE_IDS:
        rank = max(rank, _band_rank(PRIORITY_HIGH))
    if issue.rule_id == "EXTERNAL_WORKBOOK_REFERENCE":
        rank = max(rank, _band_rank(PRIORITY_HIGH))
    if issue.rule_id == "FORMULA_PATTERN_ANOMALY":
        size = _formula_cluster_size(issue)
        if size >= 5:
            rank = max(rank, _band_rank(PRIORITY_HIGH))
        elif size >= 3:
            rank = max(rank, _band_rank(PRIORITY_MEDIUM))
    if issue.rule_id == "HARDCODED_NUMERIC_CONSTANT":
        rank = max(rank, _band_rank(PRIORITY_MEDIUM))

    for factor in factors:
        if factor.code == "very_hidden_sheet":
            rank = _bump_rank(rank, 2)
        elif factor.code in {
            "hidden_sheet",
            "output_like_sheet",
            "external_workbook_dependency",
            "risky_volatile_functions",
            "formula_cluster_moderate",
        }:
            rank = _bump_rank(rank, 1)
        elif factor.code == "formula_cluster_large":
            rank = _bump_rank(rank, 2)
        elif factor.code == "isolated_low_risk_volatile":
            rank = _band_rank(PRIORITY_LOW)

    if sheet_state in {"hidden", "veryHidden"} and issue.rule_id in STRUCTURAL_RULE_IDS:
        rank = _band_rank(PRIORITY_CRITICAL)

    return replace(
        issue,
        priority=_rank_to_band(rank),
        impact_factors=tuple(sorted(factors, key=lambda factor: (factor.code, factor.explanation))),
    )


def _collect_impact_factors(issue: Issue, sheet_state: str) -> list[ImpactFactor]:
    factors: list[ImpactFactor] = []

    if sheet_state == "hidden":
        factors.append(
            ImpactFactor(
                code="hidden_sheet",
                explanation=(
                    "Issue is on a hidden worksheet, so reviewers can miss it "
                    "during a walkthrough."
                ),
            )
        )
    elif sheet_state == "veryHidden":
        factors.append(
            ImpactFactor(
                code="very_hidden_sheet",
                explanation=(
                    "Issue is on a very hidden worksheet, which is easy to "
                    "overlook without structure review."
                ),
            )
        )

    if _output_like_sheet(issue.evidence.sheet):
        factors.append(
            ImpactFactor(
                code="output_like_sheet",
                explanation=(
                    "Worksheet name suggests an output or summary surface "
                    "where defects are user-visible."
                ),
            )
        )

    if issue.rule_id == "EXTERNAL_WORKBOOK_REFERENCE":
        factors.append(
            ImpactFactor(
                code="external_workbook_dependency",
                explanation=(
                    "Formula depends on another workbook file path, version, "
                    "and availability."
                ),
            )
        )

    if issue.rule_id == "HARDCODED_NUMERIC_CONSTANT":
        factors.append(
            ImpactFactor(
                code="hardcoded_numeric_constant",
                explanation=(
                    "Control or assumption values are embedded directly in a "
                    "formula instead of labeled inputs."
                ),
            )
        )

    size = _formula_cluster_size(issue)
    if size >= 5:
        factors.append(
            ImpactFactor(
                code="formula_cluster_large",
                explanation=(
                    "Copied formula family has at least five neighboring cells, "
                    "so a local anomaly affects a wide block."
                ),
            )
        )
    elif size >= 3:
        factors.append(
            ImpactFactor(
                code="formula_cluster_moderate",
                explanation=(
                    "Copied formula family spans at least three neighboring cells, "
                    "increasing blast radius of a paste error."
                ),
            )
        )

    if issue.rule_id == "VOLATILE_FUNCTION":
        functions = _string_sequence_detail(issue.details.get("functions"))
        risky = sorted(
            function.upper()
            for function in functions
            if function.upper() in RISKY_VOLATILE_FUNCTIONS
        )
        if risky:
            factors.append(
                ImpactFactor(
                    code="risky_volatile_functions",
                    explanation="Formula uses higher-risk volatile functions: "
                    + ", ".join(risky)
                    + ".",
                )
            )
        elif _isolated_low_risk_volatile(functions):
            factors.append(
                ImpactFactor(
                    code="isolated_low_risk_volatile",
                    explanation=(
                        "Formula only uses TODAY(), which is volatile but usually "
                        "lower priority than INDIRECT/OFFSET families."
                    ),
                )
            )

    return factors


def _output_like_sheet(name: str) -> bool:
    lower = name.strip().lower()
    return any(token in lower for token in OUTPUT_LIKE_SHEET_TOKENS)


def _formula_cluster_size(issue: Issue) -> int:
    return len(_string_sequence_detail(issue.details.get("cluster_cells")))


def _string_sequence_detail(value: object) -> list[str]:
    if isinstance(value, (list, tuple)):
        return [item for item in value if isinstance(item, str)]
    return []


def _isolated_low_risk_volatile(functions: list[str]) -> bool:
    if not functions:
        return False
    return all(
        function.upper() in LOW_RISK_VOLATILE_FUNCTIONS
        and function.upper() not in RISKY_VOLATILE_FUNCTIONS
        for function in functions
    )


def _severity_base_rank(severity: str) -> int:
    if severity == "high":
        return _band_rank(PRIORITY_HIGH)
    if severity == "low":
        return _band_rank(PRIORITY_LOW)
    return _band_rank(PRIORITY_MEDIUM)


def _band_rank(band: str) -> int:
    if band == PRIORITY_CRITICAL:
        return 4
    if band == PRIORITY_HIGH:
        return 3
    if band == PRIORITY_MEDIUM:
        return 2
    return 1


def _rank_to_band(rank: int) -> str:
    if rank >= _band_rank(PRIORITY_CRITICAL):
        return PRIORITY_CRITICAL
    if rank >= _band_rank(PRIORITY_HIGH):
        return PRIORITY_HIGH
    if rank >= _band_rank(PRIORITY_MEDIUM):
        return PRIORITY_MEDIUM
    return PRIORITY_LOW


def _bump_rank(rank: int, delta: int) -> int:
    return min(rank + delta, _band_rank(PRIORITY_CRITICAL))
