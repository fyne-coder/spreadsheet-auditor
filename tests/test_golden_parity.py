from __future__ import annotations

import json
from pathlib import Path

import pytest

from spreadsheet_auditor.audit import audit_workbook

WORKBOOKS_DIR = Path(__file__).resolve().parent / "fixtures" / "workbooks"
GOLDEN_DIR = Path(__file__).resolve().parent / "fixtures" / "golden"
NORMALIZED_WORKBOOK_PATH = "<workbook>"


def _canonical_json(workbook_path: Path) -> bytes:
    report = audit_workbook(workbook_path).to_dict()
    report["workbook_path"] = NORMALIZED_WORKBOOK_PATH
    return (json.dumps(report, indent=2, sort_keys=True) + "\n").encode("utf-8")


def _fixture_workbook_paths() -> list[Path]:
    return sorted(WORKBOOKS_DIR.glob("*"))


@pytest.mark.parametrize("workbook_path", _fixture_workbook_paths(), ids=lambda path: path.stem)
def test_fixture_matches_committed_golden(workbook_path: Path) -> None:
    golden_path = GOLDEN_DIR / f"{workbook_path.stem}.json"
    assert golden_path.is_file(), f"missing golden for fixture {workbook_path.name}"

    expected = golden_path.read_bytes()
    actual = _canonical_json(workbook_path)
    assert actual == expected, (
        f"golden drift for {workbook_path.name}; "
        "run `make regenerate-goldens` and commit updated fixtures"
    )


def test_lexical_cell_ordering_puts_b10_before_b2() -> None:
    golden_path = GOLDEN_DIR / "lexical_cell_ordering.json"
    report = json.loads(golden_path.read_text(encoding="utf-8"))
    cells = [issue["evidence"]["cell"] for issue in report["issues"]]
    assert cells.index("B10") < cells.index("B2")


def test_hidden_fixture_includes_visible_hidden_and_very_hidden_states() -> None:
    golden_path = GOLDEN_DIR / "hidden_and_very_hidden.json"
    report = json.loads(golden_path.read_text(encoding="utf-8"))
    states = {sheet["name"]: sheet["state"] for sheet in report["sheets"]}
    assert states["Visible"] == "visible"
    assert states["HiddenSheet"] == "hidden"
    assert states["Vault"] == "veryHidden"
