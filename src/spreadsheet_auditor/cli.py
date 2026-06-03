from __future__ import annotations

import argparse
import json
from pathlib import Path

from spreadsheet_auditor.audit import audit_workbook
from spreadsheet_auditor.review_pack import render_review_pack_html


def main(argv: list[str] | None = None) -> int:
    parser = _build_parser()
    args = parser.parse_args(argv)

    if args.command == "scan":
        report = audit_workbook(args.workbook)
        wrote_output = False

        if args.output:
            payload = json.dumps(report.to_dict(), indent=2, sort_keys=True)
            Path(args.output).write_text(payload + "\n", encoding="utf-8")
            wrote_output = True

        if args.review_pack:
            Path(args.review_pack).write_text(
                render_review_pack_html(report),
                encoding="utf-8",
            )
            wrote_output = True

        if not wrote_output:
            payload = json.dumps(report.to_dict(), indent=2, sort_keys=True)
            print(payload)

        return 0

    parser.error("missing command")
    return 2


def _build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        prog="spreadsheet-auditor",
        description="Run static checks against Excel workbooks.",
    )
    subparsers = parser.add_subparsers(dest="command")

    scan = subparsers.add_parser("scan", help="scan a workbook and emit reports")
    scan.add_argument("workbook", help="path to a .xlsx or .xlsm workbook")
    scan.add_argument("--output", "-o", help="write JSON report to this path")
    scan.add_argument(
        "--review-pack",
        help="write HTML review pack to this path",
    )

    return parser


if __name__ == "__main__":
    raise SystemExit(main())

