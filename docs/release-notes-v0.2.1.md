# Spreadsheet Auditor v0.2.1

Date: 2026-06-03

## Release Type

Source prerelease. Signed and notarized desktop binaries are not attached yet.

## Shipped Changes

- Added README screenshots captured from the Wails desktop app using the
  synthetic `combined_risky.xlsx` demo workbook.
- Documented the analyst flow shown in those screenshots: workbook triage,
  issue details, export options, and optional AI-assistant handoff.
- Bumped source package metadata from `0.2.0` to `0.2.1`.

## Verification

Run from the sanitized v0.2.1 release worktree before tagging:

```bash
python3 - <<'PY'
from pathlib import Path
readme = Path("README.md").read_text()
for rel in [
    "docs/screenshots/desktop-overview.png",
    "docs/screenshots/desktop-issue-detail.png",
    "docs/screenshots/desktop-export-options.png",
    "docs/screenshots/desktop-ai-handoff.png",
]:
    if rel not in readme:
        raise SystemExit(f"missing README link: {rel}")
    if not Path(rel).is_file():
        raise SystemExit(f"missing screenshot file: {rel}")
print("README screenshot links verified")
PY
make check
make desktop-build
make package-smoke-mac
cd desktop/frontend && npm audit --audit-level=moderate
git diff --check
bash /Users/arthurlee/src/assistant/scripts/coding_check.sh /tmp/spreadsheet-auditor-v021.Gu7LjZ
```

## Known Risks

- This remains a source prerelease. The macOS app still needs Developer ID
  signing, notarization, stapling, and extracted-artifact smoke testing before
  external analyst testers should use a packaged binary.
- `make package-smoke-mac` confirms the local app bundle version and ad-hoc code
  signature, and is expected to report Gatekeeper rejection and no stapled
  notarization ticket until release signing is completed.
- The v0.2.0 security report still applies until the Go toolchain and hardening
  follow-ups are addressed.
