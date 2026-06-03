# Go and Wails Conversion Evaluation

Date: 2026-06-02

## Summary

This is a historical planning note from before the Go/Wails implementation
landed. The current branch followed the recommended direction: keep Python as
the parity oracle, port the analyzer to Go with Excelize, keep the Go CLI and
Wails service on the same `AuditReport` model, and verify parity through
committed fixtures. For the current implementation contract, see
`docs/architecture.md` and `docs/parity-contract.md`.

The desktop target is viable, but the safest path is not a blind rewrite. Keep the
existing Python analyzer as the parity oracle, spike the Go parser behavior against
the existing synthetic tests and ignored public smoke workbooks, then move the Wails
desktop app onto the Go analyzer only after report parity is proven.

Recommended migration:

1. Preserve the current JSON report contract.
2. Port the analyzer core to Go with Excelize.
3. Keep a Go CLI and Wails service on the same `AuditReport` model.
4. Use the Python implementation as a golden-output comparator until Go parity is
   stable.

## Why Wails Fits

Wails is a good fit for this repo's local-first posture because the analyzer can
run as compiled Go code behind a web UI without a local HTTP server or cloud upload.
The frontend can call exported Go service methods directly, and generated TypeScript
bindings can keep the UI/report contract explicit.

As of 2026-06-02, Wails v2 is the stable track and Wails v3 is alpha:

- https://github.com/wailsapp/wails
- https://wails.io/docs/howdoesitwork/
- https://v3.wails.io/concepts/architecture/

Use Wails v2 for an MVP desktop app unless there is a specific v3 feature worth
accepting alpha churn.

## Parser Candidate

Excelize is the leading Go parser candidate:

- https://github.com/xuri/excelize
- https://pkg.go.dev/github.com/xuri/excelize/v2
- https://xuri.me/excelize/en/cell.html
- https://xuri.me/excelize/en/workbook.html

It supports `.xlsx` and `.xlsm`, exposes cell formulas, workbook defined names,
sheet visibility, rows, tables, and other metadata surfaces. That maps well to the
current product direction. The main unknown is formula-tokenizer parity. The Python
code currently relies on `openpyxl.formula.Tokenizer` for numeric constants and
formula-pattern normalization, so the Go port either needs an equivalent parser or a
small tested lexical tokenizer scoped to the existing rules.

## Current Python Surfaces To Port

Core analyzer:

- `src/spreadsheet_auditor/audit.py`
  - supported suffix validation
  - static workbook open with formulas preserved
  - sheet inventory
  - formula walk
  - hardcoded numeric constants
  - volatile functions
  - broken `#REF!`
  - whole-column ranges
  - external workbook references
  - formula-pattern anomaly dispatch

Data model:

- `src/spreadsheet_auditor/models.py`
  - rule registry
  - issue evidence
  - issue records
  - sheet summaries
  - deterministic `AuditReport` serialization

Formula patterns:

- `src/spreadsheet_auditor/formula_pattern.py`
  - relative reference normalization
  - row/column consecutive clusters
  - exactly-one-outlier rule

Exports:

- `src/spreadsheet_auditor/review_pack.py`
  - HTML review-pack rendering
  - escaping workbook-provided strings

## Proposed Go Layout

Keep the analyzer independent from Wails:

```text
cmd/spreadsheet-auditor/
  main.go
internal/audit/
  audit.go
  rules.go
internal/model/
  report.go
internal/formula/
  normalize.go
  tokenizer.go
internal/reviewpack/
  html.go
desktop/
  main.go
  app.go
  frontend/
```

This keeps three product paths open:

- command-line scanner for automation
- Wails desktop app for local review
- future service or batch runner using the same analyzer package

## Wails Service Shape

The Wails backend should expose a small service, not UI-specific analyzer logic:

```go
type AuditService struct{}

func (s *AuditService) ScanWorkbook(path string) (*model.AuditReport, error)
func (s *AuditService) RenderReviewPack(path string) (string, error)
func (s *AuditService) SaveReviewPack(path string, outputPath string) error
```

The UI should own file picking, issue filtering, severity/category views, and export
actions. The Go analyzer should stay deterministic, read-only, and independent of the
desktop shell.

## Conversion Options

### Option A: Wails Shell Calling Python CLI

Fastest demo path. Wails calls the current `spreadsheet-auditor scan` command and
loads JSON into the frontend.

Pros:

- Minimal analyzer risk.
- UI can start immediately.
- Existing tests and smoke evidence remain authoritative.

Cons:

- Packaging Python and dependencies into a desktop app adds distribution friction.
- Process execution and path handling become product code.
- It does not achieve the goal of a Go analyzer.

Use only as a short-lived prototype.

### Option B: Full Go Analyzer Port Before UI

Best long-term product architecture. Port the analyzer first, verify parity, then
add Wails.

Pros:

- Single compiled backend.
- Cleaner packaging.
- Easier Wails integration.
- Analyzer remains reusable outside desktop.

Cons:

- Formula tokenizer behavior is the main schedule risk.
- Must rebuild deterministic JSON behavior exactly enough for downstream stability.
- Requires a stronger parity harness before deleting or deprecating Python.

This is the recommended target.

### Option C: Parallel Go Analyzer Plus Wails Prototype

Pragmatic migration. Build Wails around a Go service contract, initially allow the
service to call Python in development, then swap in the Go analyzer after parity.

Pros:

- UI and Go port can progress in parallel.
- Service contract stabilizes early.
- Python remains the fallback oracle.

Cons:

- Temporary complexity.
- Must avoid shipping a confusing hybrid runtime unless intentionally accepted.

This is the recommended execution path if speed matters.

## Parity Plan

1. Add golden JSON fixtures generated by Python for the current synthetic pytest
   workbooks.
2. Add a Go test harness that scans the same generated workbooks and compares
   normalized JSON.
3. Re-run the existing ignored public smoke workbooks from `test-output/` when
   available.
4. Require parity for:
   - summary counts
   - sheet order and visibility
   - formula cell counts
   - rule IDs
   - issue order
   - evidence sheet/cell/formula
   - details for formula-pattern anomalies
5. Only then switch Wails to the Go analyzer as the default backend.

## Specific Risks

- Excelize formula retrieval may differ from openpyxl for shared formulas, array
  formulas, tables, structured references, and unusual defined-name syntax.
- The current Python pattern normalizer depends on token types from openpyxl. A Go
  implementation needs focused tests before expanding beyond current rules.
- Excelize supports writing workbooks too, but this product must keep the default
  analyzer read-only. Avoid save/mutation code in scan paths.
- Wails v3 is alpha as of this evaluation. Use stable Wails v2 for packaging unless
  the project explicitly accepts alpha framework churn.
- Hidden versus very-hidden sheet state may need validation because some public Go
  APIs expose boolean visibility while the current report records string states.

## First Bounded Slice

Create a Go parity spike, not a full desktop app:

- Add `go.mod`.
- Add `internal/model` with Go structs matching the current JSON contract.
- Add `internal/audit` with suffix validation, workbook open, sheet inventory, formula
  count, and the simple regex-based rules.
- Add one Go test that scans a generated workbook and compares the JSON shape to the
  Python report.
- Record Excelize gaps before porting formula-pattern anomaly detection.

Acceptance criteria:

- `go test ./...` passes.
- Python `make check` still passes.
- One synthetic workbook produces matching summary, sheets, and basic issue IDs.
- No Wails scaffold is added until the analyzer package boundary is clean.
