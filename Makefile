SHELL := /bin/bash

GO_MODULES := libs/shared libs/domain

.PHONY: help fmt lint test tidy up down logs

help:
	@echo "Available targets:"
	@echo "  fmt   - format all Go modules"
	@echo "  lint  - run revive on all Go files"
	@echo "  test  - run tests for all Go modules"
	@echo "  tidy  - go mod tidy for all modules"
	@echo "  up    - start infrastructure"
	@echo "  down  - stop infrastructure"
	@echo "  logs  - tail infrastructure logs"

fmt:
	@set -euo pipefail; \
	for mod in $(GO_MODULES); do \
		echo "Formatting $$mod"; \
		(cd $$mod && gofmt -w $$(find . -type f -name '*.go')); \
	done

lint:
	@set -euo pipefail; \
	if ! command -v revive >/dev/null 2>&1; then \
		echo "revive is not installed. Run: go install github.com/mgechev/revive@latest"; \
		exit 1; \
	fi; \
	for mod in $(GO_MODULES); do \
		echo "Linting $$mod"; \
		(cd $$mod && revive -config ../../revive.toml ./...); \
	done

test:
	@set -euo pipefail; \
	for mod in $(GO_MODULES); do \
		echo "Testing $$mod"; \
		(cd $$mod && go test ./...); \
	done

tidy:
	@set -euo pipefail; \
	for mod in $(GO_MODULES); do \
		echo "Tidying $$mod"; \
		(cd $$mod && go mod tidy); \
	done

up:
	docker compose up -d postgres redis nats minio minio-init

down:
	docker compose down -v

logs:
	docker compose logs -f postgres redis nats minio
