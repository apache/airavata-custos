SHELL := /bin/bash
.PHONY: help gen-api build verify-no-drift test integration-test verify integration-test-amie integration-test-amie-down

help:
	@echo "Targets:"
	@echo "  gen-api                     regenerate every OpenAPI spec from swag annotations"
	@echo "  build                       gen-api, then go build ./..."
	@echo "  verify-no-drift             gen-api, then fail if any *.openapi.yaml changed"
	@echo "  test                        go test ./..."
	@echo "  integration-test            full integration suite (AMIE for now)"
	@echo "  verify                      verify-no-drift + go vet + test"
	@echo "  integration-test-amie       run the AMIE connector integration suite end-to-end"
	@echo "  integration-test-amie-down  tear down the AMIE test stack only"

gen-api:
	go generate ./...

build: gen-api
	go build -o custos ./cmd/server
	go build ./...

verify-no-drift: gen-api
	@if [ -n "$$(git status --porcelain api/ connectors/*/api/ 2>/dev/null)" ]; then \
		echo "OpenAPI drift detected — run 'make gen-api' and commit the result."; \
		git status --porcelain api/ connectors/*/api/; \
		exit 1; \
	fi

test:
	go test ./...

verify: verify-no-drift
	go vet ./...
	$(MAKE) test

integration-test: integration-test-amie

integration-test-amie:
	bash scripts/run-amie-integration-tests.sh

integration-test-amie-down:
	$(MAKE) -C dev-ops/local-amie down
