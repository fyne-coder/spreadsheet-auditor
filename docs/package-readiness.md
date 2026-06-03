# Package Readiness

Date: 2026-06-02

## Purpose

Make the desktop build diagnosable before external testing. The release checks
must identify app identity, code-signing state, Gatekeeper posture, stapler
state, and the exact version intended for GitHub release assets.

## Current Status

| Area | Status | Evidence / Gate |
| --- | --- | --- |
| Local dev build | Ready for Arthur testing | `cd desktop && wails dev` |
| macOS packaged build | Ready for local smoke | `make desktop-build` |
| macOS package smoke | Added | `make package-smoke-mac` reports bundle ID, version, signing, Gatekeeper, and stapler state |
| macOS signing check | Added | `make signing-check-mac` verifies `codesign`, entitlements visibility, `spctl`, and stapler state |
| macOS Developer ID signing | Not ready | Requires Developer ID Application certificate and hardened-runtime signing step |
| macOS notarization | Not ready | `make notarization-preflight-mac` checks notary tooling and credentials only |
| Windows packaged build | Not smoke-tested | Needs Windows machine or runner |
| Windows signing | Not ready | Requires Authenticode/MSIX decision and publisher certificate |

## macOS Commands

Dry-run commands are safe when no packaged app exists:

```bash
make package-smoke-mac-dry-run
make signing-check-mac-dry-run
make notarization-preflight-mac-dry-run
```

Local packaged smoke:

```bash
make desktop-build
make package-smoke-mac
```

The app is not externally certified until a Developer ID signed artifact passes
these checks. `spctl` and stapler validation are meaningful release gates only
after Developer ID signing and notarization are in place; unsigned local builds
may fail them even when they are acceptable for Arthur-only development.

```bash
make signing-check-mac
xcrun stapler validate "desktop/build/bin/Spreadsheet Auditor.app"
spctl --assess --type execute --verbose=4 "desktop/build/bin/Spreadsheet Auditor.app"
```

## macOS Developer ID / Notarization Inputs

Needed before external testers:

- Apple Developer Program membership for the publisher.
- Developer ID Application certificate installed in the build keychain.
- Hardened-runtime signing command or Wails-compatible signing configuration.
- Notarization credentials, either:
  - `SPREADSHEET_AUDITOR_APPLE_NOTARY_PROFILE` for a stored `notarytool` profile, or
  - `APPLE_ID`, `APPLE_TEAM_ID`, and `APPLE_APP_PASSWORD`.
- A release artifact policy: zip or dmg, signed before notarization and stapled
  after notarization succeeds.
- A clean extraction smoke of the final release zip from `/tmp`.

## Release Artifact Policy

- Keep source code and build metadata in git.
- Keep built app bundles and zip/dmg artifacts out of git.
- Attach signed, stapled desktop artifacts to GitHub Releases.
- Mark `v0.2.0` as a prerelease until the macOS artifact is signed, notarized,
  stapled, and smoke-tested from the extracted release zip.

## Readiness Rules

- Arthur-only local testing can use unsigned or ad-hoc signed builds.
- External macOS testers require Developer ID signing and notarization.
- External Windows testers require a signed distribution path or an explicit
  documented exception for a private development build.
