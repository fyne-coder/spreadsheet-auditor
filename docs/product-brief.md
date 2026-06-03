# Product Brief

## Positioning

Spreadsheet Auditor is a local-first review assistant for inherited Excel workbooks. It helps reviewers understand workbook structure, find deterministic risk signals, and produce sign-off-ready review artifacts.

It should not compete first as a generic spreadsheet AI copilot. Microsoft and Google already own much of the formula-writing and natural-language spreadsheet assistant surface. The differentiated wedge is deterministic workbook review and sign-off readiness.

## Target Users

- FP&A and finance model owners reviewing forecast, budget, close, and board workbooks
- Internal and external auditors checking workbook integrity and evidence quality
- Consultants and fractional CFO teams triaging unfamiliar client files
- Data and operations teams using Excel as last-mile business logic

## Current Promise

Open a workbook, find the highest-risk issues quickly, explain why each issue matters, and export a concise review pack.

## Current App

- Workbook map: visible, hidden, and very-hidden sheets, used ranges, formula counts, and static workbook metadata exposed by the parser.
- Formula and structure review: hardcoded constants, volatile functions, broken references, Excel error sentinels, external workbook references, whole-column ranges, and formula-pattern anomalies.
- Analyst routing: deterministic priority bands and impact factors to help reviewers decide what to inspect first.
- Review readiness: filterable issue table, issue detail drawer, HTML/CSV exports, and manager-readable review packs.
- Optional AI handoff: a copy/export evidence package for the user's chosen assistant, with paste-back citation validation.
- Privacy: local processing by default. The app does not upload workbooks or call an AI provider.

## Not Yet In The App

- Workbook comparison/change review between versions.
- Automated workbook fixes or formula rewrites.
- Suppression workflow and policy packs.
- Full dependency graph or live recalculation.
- Direct ChatGPT/OpenAI/Claude/Gemini integration.
- Signed/notarized desktop distribution for external testers.

## Explicit Non-Goals

- Autonomous formula rewriting
- Automatic workbook fixing
- Macro execution
- External link refresh
- Cloud upload as the default path
- Formula recalculation or numeric correctness certification
- Google Sheets parity
- Enterprise policy workflow

## Product KPI

The key KPI is not total scan count. It is the share of scans where users accept, act on, and share the flagged issues.
