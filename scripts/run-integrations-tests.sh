#!/usr/bin/env bash
# Run the SLURM Association-Mapper integration tests.
#
# Usage:
#   scripts/run-integrations-tests.sh                 # run all integration tests in operations/
#   scripts/run-integrations-tests.sh -run TestFoo    # forward extra flags to `go test`

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"

# Mint a fresh SLURM JWT via the local-slurm Makefile target.
# `make token` prints e.g. `SLURM_JWT=eyJhbGciOi...` — strip the prefix.
echo "==> minting SLURM JWT via 'make token'"
TOKEN_LINE="$(make -s -C dev-ops/local-slurm token | grep -E '^SLURM_JWT=' | tail -n1)"
if [[ -z "${TOKEN_LINE}" ]]; then
    echo "ERROR: 'make token' did not produce a SLURM_JWT=... line" >&2
    exit 1
fi
export TEST_SLURM_TOKEN="${TOKEN_LINE#SLURM_JWT=}"
echo "==> TEST_SLURM_TOKEN set (${#TEST_SLURM_TOKEN} chars)"

# go test -tags integration -v -count=1 \
#    ./connectors/SLURM/Association-Mapper/internal/operations/... \
#    "$@"


go test -tags integration -v -count=1 \
  ./connectors/SLURM/Association-Mapper/internal/operations/accounts.go \
  ./connectors/SLURM/Association-Mapper/internal/operations/associations.go \
  ./connectors/SLURM/Association-Mapper/internal/operations/client.go \
  ./connectors/SLURM/Association-Mapper/internal/operations/tres.go \
  ./connectors/SLURM/Association-Mapper/internal/operations/types.go \
  ./connectors/SLURM/Association-Mapper/internal/operations/integration_common.go \
  ./connectors/SLURM/Association-Mapper/internal/operations/associations_integration_test.go