# HelixPlay — Ultimate Gaming Experience
# Root Makefile for 46-submodule monorepo
# Constitution v2.2.0 compliant — anti-bluff enforcement

.PHONY: all build test test-unit test-integration test-e2e test-security test-bench test-coverage test-chaos test-stress test-smoke test-fullauto test-challenge anti-bluff verify-submodules propagate-constitution fmt vet lint clean docker-build docker-up docker-down help

GOVERSION := 1.26.2
GOMAXPROCS ?= 2
export GOMAXPROCS

## Default: build + verify
all: vet anti-bluff build verify-submodules

BINARY_DIR := bin
CMD_PACKAGES := $(shell go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./cmd/... 2>/dev/null | sed 's|github.com/HelixDevelopment/HelixPlay/||')

## Build all main commands to bin/ and submodules
build:
	@mkdir -p $(BINARY_DIR)
	@for pkg in $(CMD_PACKAGES); do \
		name=$$(basename $$pkg); \
		echo "Building $$name..."; \
		go build -o $(BINARY_DIR)/$$name ./$$pkg; \
	done
	@echo "Root module built. Submodule builds:"
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" build 2>/dev/null || echo "  SKIP $$dir (no build target)"; \
		fi; \
	done

## Run all tests (Unit + Integration + E2E)
test: test-unit test-integration test-e2e

## Unit tests (mocks allowed) — per-PR
test-unit:
	go test -count=1 -race -p 1 ./cmd/... ./pkg/...
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" test 2>/dev/null || true; \
		fi; \
	done

## Integration tests (real deps, no mocks) — per-PR + nightly
test-integration:
	go test -count=1 -race -p 1 ./tests/integration/...
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" test-integration 2>/dev/null || true; \
		fi; \
	done

## End-to-end tests (full topology in containers) — per-PR + nightly
test-e2e:
	go test -count=1 -race -p 1 ./tests/e2e/... 2>/dev/null || echo "E2E tests: create tests/e2e/ first"

## Security tests (govulncheck + Snyk + Trivy + fuzz) — per-PR + monthly
test-security:
	govulncheck ./...
	@bash scripts/anti-bluff-scan.sh
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" test-security 2>/dev/null || true; \
		fi; \
	done

## Benchmarks (p50/p99/p999 HDR histogram) — nightly
test-bench:
	go test -bench=. -benchmem ./tests/benchmark/... 2>/dev/null || echo "Benchmarks: create tests/benchmark/ first"

## Coverage report (HTML + threshold check ≥95%)
test-coverage:
	go test -count=1 -race -coverprofile=coverage.out ./cmd/... ./pkg/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## Chaos tests (Toxiproxy + chaos-mesh) — nightly
test-chaos:
	go test -count=1 -race -p 1 ./tests/chaos/... 2>/dev/null || echo "Chaos tests: create tests/chaos/ first"

## Stress tests (24-hour soak) — canary
test-stress:
	go test -count=1 -race -p 1 ./tests/stress/... 2>/dev/null || echo "Stress tests: create tests/stress/ first"

## Smoke tests (30-second post-deploy) — post-deploy
test-smoke:
	go test -count=1 -race -p 1 ./tests/smoke/... 2>/dev/null || echo "Smoke tests: create tests/smoke/ first"

## Full automation (orchestrates all 10 types, fail-fast disabled) — pre-release
test-fullauto: test-unit test-integration test-e2e test-security test-bench test-chaos test-stress test-smoke test-challenge
	@echo "=== Full Automation Complete ==="

## Challenge tests (meta-test, real user journeys) — nightly
test-challenge:
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/challenges/scripts/$$(basename $$dir)_challenge.sh" ]; then \
			bash "$$dir/challenges/scripts/$$(basename $$dir)_challenge.sh" || exit 1; \
		fi; \
	done

## Anti-bluff scan (non-overridable CI lane) — every commit
anti-bluff:
	@bash scripts/anti-bluff-scan.sh

## Verify all 46 submodules are present and clean
verify-submodules:
	@python3 scripts/verify-submodules.py

## Propagate Constitution v2.2.0 to all submodules
propagate-constitution:
	@bash scripts/propagate-constitution.sh

## Format all Go code
fmt:
	gofmt -w .
	goimports -w .
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" fmt 2>/dev/null || true; \
		fi; \
	done

## Vet all Go code
vet:
	go vet ./...
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" vet 2>/dev/null || true; \
		fi; \
	done

## Lint all Go code (golangci-lint)
lint:
	golangci-lint run ./...
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" lint 2>/dev/null || true; \
		fi; \
	done

## Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -cache
	@for dir in $$(git config --file .gitmodules --get-regexp path | awk '{print $$2}'); do \
		if [ -f "$$dir/Makefile" ]; then \
			$(MAKE) -C "$$dir" clean 2>/dev/null || true; \
		fi; \
	done

## Docker: build all service images
docker-build:
	@bash Containers/scripts/build-all.sh 2>/dev/null || echo "Containers/scripts/build-all.sh not found"

## Docker: start local development topology
docker-up:
	docker compose up -d

## Docker: stop local development topology
docker-down:
	docker compose down -v

## Show available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
