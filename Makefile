SHELL := /bin/bash
.PHONY: help integration-test-amie integration-test-amie-down

help:
	@echo "Targets:"
	@echo "  integration-test-amie       run the AMIE connector integration suite end-to-end"
	@echo "  integration-test-amie-down  tear down the AMIE test stack only"

integration-test-amie:
	bash scripts/run-amie-integration-tests.sh

integration-test-amie-down:
	$(MAKE) -C dev-ops/local-amie down
