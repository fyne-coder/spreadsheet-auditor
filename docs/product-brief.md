# Product Brief

## Positioning

Spreadsheet Auditor is a local-first review assistant for inherited Excel workbooks. It helps reviewers understand workbook structure, find deterministic risk signals, and produce sign-off-ready review artifacts.

It should not compete first as a generic spreadsheet AI copilot. Microsoft and Google already own much of the formula-writing and natural-language spreadsheet assistant surface. The differentiated wedge is deterministic workbook review and sign-off readiness.

## Target Users

- FP&A and finance model owners reviewing forecast, budget, close, and board workbooks
- Internal and external auditors checking workbook integrity and evidence quality
- Consultants and fractional CFO teams triaging unfamiliar client files
- Data and operations teams using Excel as last-mile business logic

## v0.1 Promise

Open a workbook, find the highest-risk issues quickly, explain why each issue matters, and export a concise review pack.

## v0.1 Must Haves

- Workbook map: sheets, hidden content, named ranges, formulas, values, links, and connections where available from static metadata.
- Formula linting: hardcoded constants, inconsistent formulas, broken references, suspicious blanks/errors, volatile functions, risky constructs, whole-column patterns, and unusual complexity.
- Static performance profiler: volatile functions, large range patterns, whole-column references, repeated expensive lookups, used-range bloat, and external-link sprawl.
- Change review: compare versions and classify formula, value, format, link, and named-range changes.
- Review readiness: severity scoring, comments, suppressions, exports, and manager-readable summary.
- Explainability: exact cells, rule triggered, reason, and suggested remediation.
- Privacy: local processing by default.

## Explicit Non-Goals For v0.1

- Autonomous formula rewriting
- Macro execution
- External link refresh
- Cloud upload as the default path
- Google Sheets parity
- Enterprise policy workflow

## Product KPI

The key KPI is not total scan count. It is the share of scans where users accept, act on, and share the flagged issues.

