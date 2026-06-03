#!/usr/bin/env bash
set -euo pipefail

APP_PATH="desktop/build/bin/Spreadsheet Auditor.app"
DRY_RUN=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --app)
      APP_PATH="${2:?--app requires a path}"
      shift 2
      ;;
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

EXE_PATH="$APP_PATH/Contents/MacOS/spreadsheet-auditor-desktop"

if [[ "$DRY_RUN" == "1" ]]; then
  cat <<EOF
signing check dry-run
app: $APP_PATH
bundle id: /usr/libexec/PlistBuddy -c 'Print CFBundleIdentifier' '$APP_PATH/Contents/Info.plist'
short version: /usr/libexec/PlistBuddy -c 'Print CFBundleShortVersionString' '$APP_PATH/Contents/Info.plist'
verify: codesign --verify --deep --strict --verbose=2 '$APP_PATH'
details: codesign -dv --verbose=4 '$EXE_PATH'
entitlements: codesign -d --entitlements :- '$APP_PATH'
gatekeeper: spctl --assess --type execute --verbose=4 '$APP_PATH'
stapler: xcrun stapler validate '$APP_PATH'
EOF
  exit 0
fi

if [[ "$(uname -s)" != "Darwin" ]]; then
  echo "signing check is macOS-only" >&2
  exit 2
fi

if [[ ! -x "$EXE_PATH" ]]; then
  echo "packaged executable not found: $EXE_PATH" >&2
  exit 1
fi

echo "signing check"
echo "app: $APP_PATH"
echo "bundle id: $(/usr/libexec/PlistBuddy -c 'Print CFBundleIdentifier' "$APP_PATH/Contents/Info.plist" 2>/dev/null || echo unknown)"
echo "short version: $(/usr/libexec/PlistBuddy -c 'Print CFBundleShortVersionString' "$APP_PATH/Contents/Info.plist" 2>/dev/null || echo unknown)"
codesign --verify --deep --strict --verbose=2 "$APP_PATH"
codesign -dv --verbose=4 "$EXE_PATH" 2>&1 | sed -n 's/^/  /p'
echo "entitlements:"
codesign -d --entitlements :- "$APP_PATH" 2>&1 | sed -n 's/^/  /p' || true
spctl --assess --type execute --verbose=4 "$APP_PATH"
echo "stapler:"
xcrun stapler validate "$APP_PATH" 2>&1 | sed -n 's/^/  /p' || true
