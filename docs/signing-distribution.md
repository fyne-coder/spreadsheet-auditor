# Signing And Distribution Notes

Date: 2026-06-02

## Goal

Keep app identity, signing, notarization, and release packaging explicit before
external desktop testing. The desktop app should be distributed as source in the
GitHub repository and as signed binary artifacts attached to GitHub Releases.

## macOS

Public distribution outside the Mac App Store should use Developer ID signing,
hardened runtime, Apple notarization, and stapling. Unsigned local builds are
acceptable for Arthur-only development, but they are not representative of
default Gatekeeper behavior for external testers.

Current app identity:

- app name: `Spreadsheet Auditor`
- bundle identifier: `com.fynellc.spreadsheet-auditor`
- output executable: `spreadsheet-auditor-desktop`
- publisher: `Fyne LLC`
- version: `0.2.0`

Readiness gates:

```bash
make desktop-build
make package-smoke-mac
make signing-check-mac
make notarization-preflight-mac
```

The final macOS release artifact should be a zipped, stapled app bundle attached
to a GitHub Release, not committed to the repository.

## Windows

Windows remains a portability target, but it is not certified in v0.2. Public
Windows distribution should use either a signed installer or MSIX path. A
non-Store installer needs Authenticode signing with SHA-256 timestamping and
SmartScreen reputation planning.

## Secrets And Artifacts

Do not commit certificates, private keys, app-specific passwords, notary
profiles, notarization logs containing account identifiers, generated zip/dmg
release assets, or built app bundles.
