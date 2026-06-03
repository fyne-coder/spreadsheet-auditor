#!/usr/bin/env bash
set -euo pipefail

DRY_RUN=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

if [[ "$DRY_RUN" == "1" ]]; then
  cat <<'EOF'
notarization preflight dry-run
tool: xcrun notarytool --version
auth option 1: SPREADSHEET_AUDITOR_APPLE_NOTARY_PROFILE for a stored notarytool profile
auth option 2: APPLE_ID + APPLE_TEAM_ID + APPLE_APP_PASSWORD
artifact expectation: signed zip or dmg produced after Developer ID signing
post-notarization check: xcrun stapler validate <app-or-dmg>
EOF
  exit 0
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "notarization preflight is macOS-only" >&2
  exit 2
fi

xcrun notarytool --version >/dev/null

if [[ -n "${SPREADSHEET_AUDITOR_APPLE_NOTARY_PROFILE:-}" ]]; then
  echo "notarization profile configured: $SPREADSHEET_AUDITOR_APPLE_NOTARY_PROFILE"
  exit 0
fi

missing=()
[[ -n "${APPLE_ID:-}" ]] || missing+=("APPLE_ID")
[[ -n "${APPLE_TEAM_ID:-}" ]] || missing+=("APPLE_TEAM_ID")
[[ -n "${APPLE_APP_PASSWORD:-}" ]] || missing+=("APPLE_APP_PASSWORD")

if [[ "${#missing[@]}" -gt 0 ]]; then
  echo "notarization credentials not configured: ${missing[*]}" >&2
  echo "set SPREADSHEET_AUDITOR_APPLE_NOTARY_PROFILE or APPLE_ID/APPLE_TEAM_ID/APPLE_APP_PASSWORD" >&2
  exit 1
fi

echo "notarization credentials are configured"
