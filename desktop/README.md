# Spreadsheet Auditor Desktop

Wails v2 desktop shell around the Go analyzer. The UI is a thin React layer over
`AuditService`; workbook parsing, lint rules, and review-pack rendering stay in
`internal/audit` and `internal/reviewpack`.

## Service Contract

| Method | Purpose |
| --- | --- |
| `ScanWorkbook(path)` | Returns `*model.AuditReport` from the Go analyzer |
| `SaveExport(path, outputPath, format, exportedAtRFC3339, issueIDs)` | Writes HTML or CSV; optional `issue_id` filter |
| `RenderReviewPack(path)` | Scans and returns HTML review-pack text |
| `SaveReviewPack(path, outputPath)` | Deprecated HTML wrapper around `SaveExport` |
| `PickWorkbook()` | Native open dialog for `.xlsx` / `.xlsm` |
| `PickExportSavePath(format, defaultName)` | Native save dialog for HTML or CSV |
| `PickReviewPackSavePath(defaultName)` | Deprecated alias for HTML save dialog |

## UI

React + TypeScript + Vite frontend with Mantine controls and TanStack Table for
issue review:

- Workbook path input with browse + scan/rescan
- Summary counts and severity/category rollups
- Sortable, searchable, faceted issue table with pagination, column visibility,
  row selection, and a detail drawer
- Copy actions for issue key, rule, cell, and formula
- Export modal (Mantine) for HTML or CSV with explicit scope:
  - all issues (empty `issueIDs` to backend)
  - selected table rows (`issue_id` keys from `src/lib/issueKey.ts`)
- Shows default filename (`review-pack.html` / `review-pack.csv`), RFC3339
  exported-at at confirm time, and success/cancel/error notifications

Production assets are built to `frontend/dist/` and embedded by Wails.
Wails bindings live in `frontend/wailsjs/`.

## Frontend Commands

From the repository root (Node 20+; see `frontend/.nvmrc`):

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

`wails.json` wires `frontend:install` and `frontend:build` to `npm ci` and
`npm run build` under `desktop/frontend/`.

## Tests

```bash
go test ./desktop/...
cd desktop/frontend && npm test -- --run
```

`TestAuditServiceMatchesCLIForCombinedRiskyFixture` proves the service returns the
same canonical JSON as the Go CLI for the `combined_risky` fixture.
