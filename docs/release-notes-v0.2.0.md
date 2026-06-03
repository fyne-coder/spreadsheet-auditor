# Spreadsheet Auditor v0.2.0

Date: 2026-06-03

## Release Type

Source prerelease. Signed and notarized desktop binaries are not attached yet.

## Shipped Changes

- Added the Wails desktop auditor workbench with workbook selection, scan
  status, audit overview, priority buckets, issue filters, issue details, and
  export flows.
- Added HTML and CSV review-pack exports with owner-summary, detailed-audit, and
  issue-list workflows.
- Kept export privacy defaults local-first: workbook identity uses the basename
  by default, and full path export is explicit opt-in.
- Added analyst priority bands and impact factors for triaging deterministic
  audit findings.
- Added formula-lint improvements, stable issue IDs, Excel error sentinel
  detection, sparse scanning behavior, and additional workbook parity fixtures.
- Added the optional manual AI-assistant handoff flow: package preview,
  sheet/cell exclusions, provider-size presets, redacted evidence packet export,
  pasted response validation, and cited Understanding rendering.
- Added verified AI analysis export: saved analysis remains validated JSON, and
  the copy action emits analyst-readable Markdown.
- Added `security_best_practices_report.md` with the current security
  best-practices review and release-hardening backlog.
- Updated analyst-facing documentation so the README leads with what the app
  does, what it does not do, local-first privacy boundaries, unsigned desktop
  distribution status, and the Python contributor-only role.

## Verification

Run from the sanitized v0.2.0 release worktree before tagging:

```bash
make check
make desktop-build
make package-smoke-mac
git diff --check
bash /Users/arthurlee/src/assistant/scripts/coding_check.sh /tmp/spreadsheet-auditor-v020.uZT1kQ
```

Additional security-review evidence from
`security_best_practices_report.md`:

```bash
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
cd desktop/frontend && npm audit --audit-level=moderate
```

## Known Risks

- This is a source prerelease. The macOS app still needs Developer ID signing,
  notarization, stapling, and extracted-artifact smoke testing before external
  analyst testers should use a packaged binary.
- `make package-smoke-mac` confirms the local app bundle version and ad-hoc code
  signature, and is expected to report Gatekeeper rejection and no stapled
  notarization ticket until release signing is completed.
- `govulncheck` reported `GO-2026-5037` for the local Go `1.26.2` standard
  library used during review. Build signed binaries with a fixed Go patch
  release before distributing them.
- The security report also records hardening follow-ups for workbook resource
  budgets, older CLI/Python output permissions, Markdown copy escaping, and CI
  dependency-audit gates.
