.PHONY: check test lint regenerate-goldens verify-goldens go-test \
	desktop-frontend-install desktop-frontend-check desktop-bindings desktop-build \
	package-smoke-mac package-smoke-mac-dry-run signing-check-mac \
	signing-check-mac-dry-run notarization-preflight-mac \
	notarization-preflight-mac-dry-run

PYTHON ?= $(if $(wildcard .venv/bin/python),.venv/bin/python,python)
WAILS ?= $(shell go env GOPATH)/bin/wails

check: lint test desktop-frontend-check go-test

go-test:
	go test ./...

desktop-frontend-install:
	cd desktop/frontend && npm ci

desktop-frontend-check: desktop-frontend-install
	cd desktop/frontend && npm run typecheck
	cd desktop/frontend && npm run lint
	cd desktop/frontend && npm test -- --run
	cd desktop/frontend && npm run build

desktop-bindings:
	rm -rf desktop/frontend/node_modules
	cd desktop && $(WAILS) generate module

desktop-build: desktop-bindings
	cd desktop && $(WAILS) build -clean

package-smoke-mac:
	@./scripts/package_smoke_macos.sh

package-smoke-mac-dry-run:
	@./scripts/package_smoke_macos.sh --dry-run

signing-check-mac:
	@./scripts/signing_check_macos.sh

signing-check-mac-dry-run:
	@./scripts/signing_check_macos.sh --dry-run

notarization-preflight-mac:
	@./scripts/notarization_preflight_macos.sh

notarization-preflight-mac-dry-run:
	@./scripts/notarization_preflight_macos.sh --dry-run

lint:
	$(PYTHON) -m ruff check src tests scripts
	$(PYTHON) -m compileall -q src tests scripts

test:
	$(PYTHON) -m pytest

regenerate-goldens:
	$(PYTHON) scripts/emit_goldens.py
	UPDATE_EVIDENCE_PACKET_GOLDENS=1 go test ./internal/evidence -run TestEvidencePacketGoldenFixtures -count=1

verify-goldens:
	$(PYTHON) scripts/emit_goldens.py --verify
	go test ./internal/audit -run TestGoldenParityFixtures -count=1
	go test ./internal/evidence -run TestEvidencePacketGoldenFixtures -count=1
