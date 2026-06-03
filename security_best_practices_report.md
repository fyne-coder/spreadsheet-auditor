# Security Best Practices Report

Date: 2026-06-03

## Executive Summary

No critical or high-severity application-code vulnerabilities were found in this
pass. The current repo has strong local-first privacy defaults in the main
desktop/Go path: no network-facing server code, no macro execution path, no OS
command execution, no dangerous React HTML sinks in app code, HTML export
escaping, CSV injection prefixing, basename-by-default exports, redacted AI
handoff packets, and private `0600` permissions for desktop JSON/review-pack
exports.

The main security work left is hardening and release hygiene:

- update the Go toolchain and add routine vulnerability scanning
- add product-level workbook resource budgets before parsing untrusted files
- align older CLI/Python output permissions with the desktop/private export
  policy
- escape or neutralize Markdown control characters in copied AI analysis

Scope reviewed:

- Go analyzer, CLI, Wails service, evidence packet, review-pack export code
- React/TypeScript Wails frontend and paste-back/copy flows
- Python CLI/parity analyzer path
- CI and package metadata

Reference guidance loaded:

- `security-best-practices/references/golang-general-backend-security.md`
- `security-best-practices/references/javascript-general-web-frontend-security.md`
- `security-best-practices/references/javascript-typescript-react-web-frontend-security.md`

No Python CLI-specific reference exists in the skill directory; Python findings
below use general secure file-output and dependency hygiene principles.

## Critical Findings

None.

## High Findings

None.

## Medium Findings

### SBP-1: Go toolchain has a reachable standard-library vulnerability and CI does not run vulnerability scanning

- Rule ID: `GO-DEPLOY-001`
- Severity: Medium
- Location:
  - `go.mod:3`
  - `.github/workflows/ci.yml:11`
  - `.github/workflows/ci.yml:16`
  - local command evidence from `go version` / `govulncheck`
- Evidence:

```text
go.mod:3: go 1.24.0
.github/workflows/ci.yml:11: - uses: actions/checkout@v4
.github/workflows/ci.yml:16: - run: make check
```

Current local toolchain:

```text
go version go1.26.2 darwin/arm64
```

`govulncheck` reported:

```text
Vulnerability #1: GO-2026-5037
    Inefficient candidate hostname parsing in crypto/x509
  Standard library
    Found in: crypto/x509@go1.26.2
    Fixed in: crypto/x509@go1.26.4
...
Your code is affected by 1 vulnerability from the Go standard library.
```

- Impact: Release builds made with the affected Go patch level inherit the
  vulnerable standard library. This is a build/toolchain issue rather than a
  workbook parsing bug, but it matters for signed desktop artifacts and any CLI
  binaries distributed from this repo.
- Fix: Upgrade the build toolchain to a fixed Go patch release and add
  `govulncheck ./...` to CI or release preflight. Pin CI's Go version explicitly
  with `actions/setup-go` so release evidence does not depend on the runner
  image's default Go installation.
- Mitigation: Until CI is updated, run `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`
  before producing any signed/notarized binary.
- False positive notes: `govulncheck` also reported imported-module
  vulnerabilities that do not appear reachable. This finding is specifically
  about the reachable standard-library vulnerability and missing recurring
  vulnerability gate.

### SBP-2: Workbook parsing has library defaults but no product-level resource budget

- Rule ID: `GO-RESOURCE-001`
- Severity: Medium
- Location:
  - `internal/audit/workbook.go:32`
  - `internal/audit/sheet_scan.go:49`
  - `internal/audit/sheet_scan.go:62`
  - `src/spreadsheet_auditor/audit.py:44`
  - `src/spreadsheet_auditor/audit.py:58`
- Evidence:

```go
// internal/audit/workbook.go:32-34
file, err := excelize.OpenFile(workbookPath, excelize.Options{
    UnzipXMLSizeLimit: 0,
})
```

```go
// internal/audit/sheet_scan.go:49-63
rows, err := file.GetRows(sheetName)
...
for rowIndex, columns := range rows {
```

```python
# src/spreadsheet_auditor/audit.py:44-50
workbook = load_workbook(
    workbook_path,
    data_only=False,
    read_only=False,
    keep_vba=suffix == ".xlsm",
    keep_links=False,
)
```

Local Excelize source confirms that `0` falls back to library defaults, not no
limit:

```text
excelize.go:86-94: default unzip size limit is 16GB; default worksheet/shared-string
memory threshold is 16MB.
excelize.go:172-179: zero options are replaced with default limits.
```

- Impact: A malicious or unusually large workbook can still consume substantial
  CPU, memory, disk temp space, and time during local analysis. For a desktop app
  this is primarily local denial of service, but inherited workbooks are exactly
  the untrusted input this tool is designed to open.
- Fix: Add explicit product-level limits before and during scan, for example:
  maximum input file size, maximum total unzipped size lower than Excelize's
  16GB default, maximum sheets, maximum rows/cells/formulas scanned, and a
  cancelable/timeout-aware scan path for the desktop UI. Keep the limits visible
  in user-facing error messages.
- Mitigation: Document recommended maximum workbook size for the current release
  and avoid opening untrusted workbooks from unknown sources until scan budgets
  are enforced.
- False positive notes: Excelize does provide default zip/XML limits and the Go
  scanner uses `GetRows()` instead of declared-dimension rectangle walks. The
  issue is absence of a tighter product budget suitable for a local desktop
  security posture.

### SBP-3: Older CLI/Python output paths can write workbook-derived reports with non-private permissions

- Rule ID: `GO-FILE-001` / `PY-FILE-001`
- Severity: Medium
- Location:
  - `cmd/spreadsheet-auditor/main.go:75`
  - `cmd/spreadsheet-auditor/main.go:81`
  - `src/spreadsheet_auditor/cli.py:19`
  - `src/spreadsheet_auditor/cli.py:21`
  - `src/spreadsheet_auditor/cli.py:24`
  - `src/spreadsheet_auditor/cli.py:25`
- Evidence:

```go
// cmd/spreadsheet-auditor/main.go:75-81
if jsonTarget != "" {
    payload, err := report.CanonicalJSON()
    ...
    if err := os.WriteFile(jsonTarget, payload, 0o644); err != nil {
```

```python
# src/spreadsheet_auditor/cli.py:19-28
if args.output:
    payload = json.dumps(report.to_dict(), indent=2, sort_keys=True)
    Path(args.output).write_text(payload + "\n", encoding="utf-8")
...
if args.review_pack:
    Path(args.review_pack).write_text(
```

By contrast, the current Go review-pack/export and desktop AI handoff paths use
private file modes:

```go
// internal/reviewpack/export.go:12
const privateExportFileMode = 0o600

// desktop/audit_service.go:82,90,114
return os.WriteFile(outputPath, payload, 0o600)
```

- Impact: JSON and HTML review artifacts can include workbook paths, sheet
  names, formulas, external references, and workbook-derived evidence. On shared
  machines or permissive umasks, the older CLI/Python paths may leave those
  artifacts readable by other local users.
- Fix: Align all report/export writers on the existing private export policy.
  In Go, write JSON output with `0600` using `os.OpenFile` or `os.WriteFile`
  with mode `0o600`. In Python, use `os.open(..., 0o600)` or an equivalent
  helper instead of `Path.write_text` for report artifacts.
- Mitigation: Document that generated reports are workbook-derived work product
  and should be stored in private directories until the file mode is fixed.
- False positive notes: The desktop path and Go HTML/CSV export path already use
  private modes. This finding is limited to the Go JSON CLI output and the
  legacy/contributor Python CLI path.

## Low Findings

### SBP-4: Copied Markdown analysis does not neutralize Markdown control characters

- Rule ID: `JS-CONTENT-001`
- Severity: Low
- Location:
  - `desktop/frontend/src/lib/understandingMarkdown.ts:3`
  - `desktop/frontend/src/lib/understandingMarkdown.ts:34`
  - `desktop/frontend/src/lib/understandingMarkdown.ts:42`
  - `desktop/frontend/src/components/understanding/UnderstandingPanel.tsx:567`
- Evidence:

```ts
// desktop/frontend/src/lib/understandingMarkdown.ts:3-4
function cleanText(value: string | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim();
}
```

```ts
// desktop/frontend/src/lib/understandingMarkdown.ts:33-34
(report.workbook_purpose ?? []).map(
  (item) => `- ${cleanText(item.claim)}\n${citationText(item.citations)}`,
)
```

```ts
// desktop/frontend/src/components/understanding/UnderstandingPanel.tsx:567-570
void notifyCopy(
  "Verified AI analysis Markdown",
  understandingReportToMarkdown(acceptedReport),
)
```

- Impact: The app copies externally generated AI text into Markdown without
  escaping Markdown/HTML control characters. If that Markdown is pasted into
  Slack, GitHub, Confluence, or a document renderer, malicious or prompt-injected
  content could hide links, render images, create misleading emphasis, or trigger
  mentions. This is not DOM XSS inside the Wails app because React renders the
  same report with normal escaping; it is a downstream content-integrity issue.
- Fix: Escape Markdown syntax for all LLM-provided text fields, or copy as plain
  text with citations instead of Markdown. If Markdown formatting is desired,
  allow only app-generated Markdown structure and treat all claim/role/risk text
  as escaped text.
- Mitigation: Keep the existing caveat at the top of copied Markdown and advise
  analysts to review pasted output before sharing externally.
- False positive notes: Citation validation confirms cited IDs exist; it does
  not validate the truth or safety of arbitrary prose returned by an LLM.

### SBP-5: CI lacks explicit Node/Go setup and recurring frontend/Python dependency audit gates

- Rule ID: `SUPPLY-CHAIN-001`
- Severity: Low
- Location:
  - `.github/workflows/ci.yml:11`
  - `.github/workflows/ci.yml:12`
  - `.github/workflows/ci.yml:16`
  - `desktop/frontend/package-lock.json`
  - absence of a Python lock/audit tool in `pyproject.toml`
- Evidence:

```yaml
# .github/workflows/ci.yml:11-16
- uses: actions/checkout@v4
- uses: actions/setup-python@v5
  with:
    python-version: "3.12"
- run: python -m pip install -e ".[dev]"
- run: make check
```

Local check evidence:

```text
npm audit --audit-level=moderate
found 0 vulnerabilities

.venv/bin/python -m pip_audit
No module named pip_audit
```

- Impact: CI currently verifies tests/lint/build through `make check`, but it
  does not make dependency vulnerability scanning a recurring gate. Frontend
  dependencies are lockfile-backed, while Python dependencies are range-based and
  not audited in this repo.
- Fix: Add explicit `actions/setup-go` and `actions/setup-node` steps, then run
  `govulncheck ./...`, `npm audit --audit-level=moderate`, and either
  `pip-audit` or a documented Python dependency review gate. Consider a lockfile
  strategy for reproducible Python contributor tooling if the Python path remains
  active.
- Mitigation: Run the dependency-audit commands manually before release until CI
  is updated.
- False positive notes: `npm audit` was clean in this local pass. This finding is
  about making the check durable in CI, not about a current npm vulnerability.

## Positive Controls Observed

- No app-code hits for `dangerouslySetInnerHTML`, `innerHTML`, `eval`, dynamic
  `Function`, `postMessage`, browser storage, frontend env-secret exposure,
  `exec.Command`, HTTP listeners, pprof, or expvar.
- Go HTML review-pack rendering escapes workbook-derived fields with
  `html.EscapeString` before interpolation (`internal/reviewpack/render.go:26`,
  `internal/reviewpack/render.go:180`).
- Go CSV exports use `encoding/csv` and prefix formula-like fields to reduce CSV
  injection risk (`internal/reviewpack/csv.go:43`, `internal/reviewpack/csv.go:73`,
  `internal/reviewpack/csv.go:85`).
- Go review-pack exports and desktop AI handoff saves use `0600` private file
  permissions (`internal/reviewpack/export.go:12`, `desktop/audit_service.go:82`,
  `desktop/audit_service.go:90`, `desktop/audit_service.go:114`).
- Evidence packets normalize workbook identity to the basename before redaction
  and rebuild the citation map after exclusions/redaction
  (`internal/evidence/packet.go:27`, `internal/promptpack/bundle.go:26`,
  `internal/promptpack/bundle.go:27`, `internal/promptpack/bundle.go:28`).
- The current product documentation correctly states that the default app does
  not upload workbooks, execute macros, refresh external data, or claim signed
  external-test readiness before Developer ID signing/notarization
  (`README.md:19`, `README.md:30`, `docs/package-readiness.md:41`).

## Commands And Evidence

Commands run from `/Users/arthurlee/src/spreadsheet-auditor` unless noted:

```bash
find /Users/arthurlee/.codex/skills/security-best-practices -maxdepth 3 -type f | sort
sed -n '1,260p' /Users/arthurlee/.codex/skills/security-best-practices/references/golang-general-backend-security.md
sed -n '1,280p' /Users/arthurlee/.codex/skills/security-best-practices/references/javascript-general-web-frontend-security.md
sed -n '1,300p' /Users/arthurlee/.codex/skills/security-best-practices/references/javascript-typescript-react-web-frontend-security.md
rg --files -g 'go.mod' -g 'package.json' -g 'pyproject.toml' -g 'wails.json' -g 'README.md' -g 'AGENTS.md' -g '*.go' -g '*.ts' -g '*.tsx' -g '*.py'
sed -n '1,220p' go.mod
sed -n '1,220p' desktop/frontend/package.json
sed -n '1,220p' pyproject.toml
sed -n '1,180p' desktop/wails.json
rg -n 'dangerouslySetInnerHTML|innerHTML|outerHTML|insertAdjacentHTML|document\.write|eval\(|new Function|setTimeout\(\s*['\"']|setInterval\(\s*['\"']|window\.location|location\.href|window\.open|target=|postMessage|localStorage|sessionStorage|import\.meta\.env|process\.env|VITE_|REACT_APP_|apiKey|secret|token|password|client_secret' desktop/frontend/src desktop/frontend/index.html -S
rg -n 'os\.Exec|exec\.Command|syscall|unsafe|cgo|http\.ListenAndServe|ListenAndServeTLS|http\.Server|pprof|expvar|io\.ReadAll|ReadFile|WriteFile|Create\(|Open\(|Mkdir|Temp|filepath\.Join|filepath\.Clean|runtime\.OpenFileDialog|runtime\.SaveFileDialog|Shell|Command|StartProcess|net/http|url\.Parse|http\.Get|http\.Post' . -S -g '*.go' -g '*.py' -g '*.ts' -g '*.tsx'
rg -n 'WriteFile\(|OpenFile\(|load_workbook\(|excelize\.OpenFile|UnzipXMLSizeLimit|dangerouslySetInnerHTML|innerHTML|eval\(|exec\.Command|http\.ListenAndServe|pprof|localStorage|sessionStorage|import\.meta\.env|process\.env' --glob '!desktop/frontend/node_modules/**' --glob '!desktop/frontend/dist/**' --glob '!desktop/build/**' . -S
go version
go env GOVERSION GOTOOLCHAIN GOPATH GOSUMDB GOPROXY GONOSUMDB GOINSECURE
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
cd desktop/frontend && npm audit --audit-level=moderate
.venv/bin/python -m pip_audit --version && .venv/bin/python -m pip_audit
rg -n 'govulncheck|npm audit|pip-audit|pip_audit|GOSUMDB|GOINSECURE|GONOSUMDB|security' .github Makefile README.md docs scripts pyproject.toml go.mod desktop/frontend/package.json -S
```

Command outcomes:

- `govulncheck` exited non-zero and reported reachable `GO-2026-5037` in the
  Go standard library, fixed in Go `1.26.4`.
- `npm audit --audit-level=moderate` reported `found 0 vulnerabilities`.
- Python audit was skipped because `.venv/bin/python -m pip_audit` failed with
  `No module named pip_audit`.
