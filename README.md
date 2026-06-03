# Spreadsheet Auditor

Local-first workbook review for inherited Excel files.

The product wedge is narrow: inspect unfamiliar `.xlsx` and `.xlsm` workbooks, surface deterministic risk signals, and export review-ready artifacts. The starting opportunity report lives in [docs/opportunity-report.md](docs/opportunity-report.md).

## Current Scope

This bootstrap implements a small static analyzer:

- workbook inventory for visible, hidden, and very-hidden sheets
- formula cell counts by sheet
- deterministic lint issues for hardcoded numeric constants in formulas
- volatile function detection
- broken `#REF!` reference detection
- whole-column range pattern detection
- external workbook reference detection
- JSON report output
- HTML review-pack export for manager-readable triage

The analyzer is read-only. It does not execute macros, refresh external links, open data connections, or evaluate formulas.

## Quick Start

```bash
python3 -m venv .venv
. .venv/bin/activate
python -m pip install -e ".[dev]"
make check
```

Scan a workbook and write JSON:

```bash
spreadsheet-auditor scan path/to/workbook.xlsx --output audit-report.json
```

Export a manager-readable HTML review pack from the same scan:

```bash
spreadsheet-auditor scan path/to/workbook.xlsx --review-pack review-pack.html
```

Write both artifacts in one run:

```bash
spreadsheet-auditor scan path/to/workbook.xlsx \
  --output audit-report.json \
  --review-pack review-pack.html
```

Generated review packs are local artifacts; keep them out of version control (see `.gitignore`).

## Go CLI

The Go analyzer mirrors the Python report contract and passes the committed golden
fixtures:

```bash
go run ./cmd/spreadsheet-auditor scan tests/fixtures/workbooks/combined_risky.xlsx \
  --output /tmp/report.json \
  --review-pack /tmp/review-pack.html \
  --export-csv /tmp/review-pack.csv \
  --exported-at 2026-06-02T12:00:00Z
```

Run Go tests and golden verification:

```bash
go test ./...
make verify-goldens
```

## Desktop App (Wails v2)

The desktop shell lives under `desktop/` and calls the same Go analyzer packages as
the CLI (`internal/audit`, `internal/reviewpack`). It does not invoke Python.

Prerequisites:

- Go 1.24+
- [Wails v2.12.0](https://wails.io/) CLI installed in `GOPATH/bin` or available as `wails`

Regenerate frontend bindings after changing `AuditService` methods:

```bash
make desktop-bindings
```

Build the macOS app:

```bash
make desktop-build
```

Run in dev mode:

```bash
cd desktop && wails dev
```

The built app is written to `desktop/build/bin/`. See [desktop/README.md](desktop/README.md)
for UI behavior and service methods.

### Desktop Signing And Release

The v0.1 desktop identity is `Spreadsheet Auditor` with bundle identifier
`com.fynellc.spreadsheet-auditor` and product version `0.1.0`.

Local package smoke and certification preflight:

```bash
make desktop-build
make package-smoke-mac
make signing-check-mac
make notarization-preflight-mac
```

External macOS tester builds should be Developer ID signed, notarized, stapled,
and shipped as GitHub Release assets rather than committed binaries. See
[docs/signing-distribution.md](docs/signing-distribution.md) and
[docs/package-readiness.md](docs/package-readiness.md).

## v0.1 Direction

The v0.1 target is a deterministic local desktop analyzer for finance, audit, consulting, and operations teams that need fast workbook triage before close, board reporting, audit review, investor delivery, or client handoff.

See:

- [docs/product-brief.md](docs/product-brief.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/parity-contract.md](docs/parity-contract.md)
- [docs/package-readiness.md](docs/package-readiness.md)
