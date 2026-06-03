# Spreadsheet Auditor Desktop

This is implementation documentation for the Wails desktop shell. Analyst-facing
product guidance lives in the root `README.md`; this file is for developers
working on the desktop app.

Wails v2 desktop shell around the Go analyzer. The UI is a thin React layer over
`AuditService`; workbook parsing, lint rules, and review-pack rendering stay in
`internal/audit` and `internal/reviewpack`.

## Service Contract

| Method | Purpose |
| --- | --- |
| `ScanWorkbook(path)` | Returns `*model.AuditReport` from the Go analyzer |
| `BuildAIHandoff(path, options)` | **Preferred.** One scan; prompt text, packet JSON, bundle JSON, and structured bundle |
| `BuildEvidencePacket(path)` | Raw Slice 1 evidence packet (no redaction/exclusions) |
| `BuildPromptBundle(path, options)` | Compatibility wrapper over `BuildAIHandoff` |
| `BuildEvidencePacketJSON(path, options)` | Compatibility wrapper over `BuildAIHandoff` |
| `BuildPromptBundleJSON(path, options)` | Compatibility wrapper over `BuildAIHandoff` |
| `SaveEvidencePacket(path, outputPath, options)` | Writes redacted/excluded packet JSON |
| `SavePromptBundle(path, outputPath, options)` | Writes full prompt bundle JSON |
| `SaveUnderstandingReport(path, rawJSON, outputPath, options)` | Re-validates pasted AI analysis citations and writes verified JSON |
| `SaveExport(path, outputPath, format, exportedAtRFC3339, issueIDs, includeFullPath)` | Writes private HTML or CSV exports; optional `issue_id` filter; basename-only workbook identity unless full path is opted in |
| `RenderReviewPack(path)` | Scans and returns HTML review-pack text |
| `SaveReviewPack(path, outputPath)` | Deprecated HTML wrapper around `SaveExport` |
| `PickWorkbook()` | Native open dialog for `.xlsx` / `.xlsm` |
| `PickExportSavePath(format, defaultName)` | Native save dialog for HTML or CSV |
| `PickReviewPackSavePath(defaultName)` | Deprecated alias for HTML save dialog |
| `ValidateUnderstandingReport(path, rawJSON, options)` | Validates pasted LLM JSON citation IDs against the packet map (`citations_resolved`; `valid` is an alias) |
| `PickJSONSavePath(defaultName)` | Native save dialog for evidence packet or prompt bundle JSON |

## UI

React + TypeScript + Vite frontend with Mantine controls and TanStack Table for
issue review:

- Workbook path input with browse + scan/rescan
- Summary counts and severity/category rollups
- Sortable, searchable, faceted issue table with pagination, column visibility,
  row selection, and a detail drawer
- Copy actions for issue key, rule, cell, and formula
- Optional AI-assistant handoff panel after scan:
  - exclusion inputs for exact sheet names and `sheet!cell` citations
  - preview tabs for prompt text, manifest, findings, formula families, and
    evidence references
  - copy/save package text, prompt bundle JSON, and evidence packet JSON
  - paste-back textarea with citation validation and grounded rendering
  - save verified AI analysis as JSON and copy it as Markdown after cited
    evidence resolves
- Export modal (Mantine) for HTML or CSV with explicit scope:
  - export job choices for owner summary, detailed audit, and issue list
    workflows; owner summary and detailed audit default to HTML, while issue
    list defaults to CSV
  - all issues (empty `issueIDs` to backend)
  - selected table rows (`issue_id` keys from `src/lib/issueKey.ts`)
- Exported workbook identity uses the basename by default; the full local path
  is included only when the user opts in
- Shows default filename (`review-pack.html` / `review-pack.csv`), RFC3339
  exported-at at confirm time, and success/cancel/error notifications

Production assets are built to `frontend/dist/` and embedded by Wails.
Wails bindings live in `frontend/wailsjs/`.

## Frontend Commands

From the repository root (Node 20+; see `desktop/frontend/.nvmrc`):

```bash
make desktop-frontend-install   # npm ci in desktop/frontend
cd desktop/frontend && npm run typecheck
cd desktop/frontend && npm run lint
cd desktop/frontend && npm test -- --run
cd desktop/frontend && npm run build
```

Or run the full frontend gate:

```bash
make desktop-frontend-check
```

## Desktop Commands

```bash
make desktop-bindings   # regenerate JS bindings
make desktop-build      # production build (npm ci + vite build + wails build)
cd desktop && wails dev # Vite dev server + Wails shell
```

`make desktop-bindings` removes `desktop/frontend/node_modules` before running
Wails generation, so the next frontend check or build performs a fresh
`npm ci`.

`wails.json` wires `frontend:install` and `frontend:build` to `npm ci` and
`npm run build` under `desktop/frontend/`.

## Tests

```bash
go test ./desktop/...
cd desktop/frontend && npm test -- --run
```

`TestAuditServiceMatchesCLIForCombinedRiskyFixture` proves the service returns the
same canonical JSON as the Go CLI for the `combined_risky` fixture.
