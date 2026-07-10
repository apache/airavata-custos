#!/usr/bin/env bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

# Run the SLURM Association-Mapper integration tests.
#
# Usage:
#   scripts/run-integrations-tests.sh                 # run all integration tests in operations/
#   scripts/run-integrations-tests.sh -run TestFoo    # forward extra flags to `go test`

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${REPO_ROOT}"

#make -s -C dev-ops/local-slurm down
#make -s -C dev-ops/local-slurm build
#make -s -C dev-ops/local-slurm up

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

TOKEN_LINE2="$(make -s -C dev-ops/local-slurm token2 | grep -E '^SLURM_JWT=' | tail -n1)"
if [[ -z "${TOKEN_LINE2}" ]]; then
    echo "ERROR: 'make token2' did not produce a SLURM_JWT=... line" >&2
    exit 1
fi

export TEST_SLURM_TOKEN2="${TOKEN_LINE2#SLURM_JWT=}"
echo "==> TEST_SLURM_TOKEN2 set (${#TEST_SLURM_TOKEN2} chars)"


TOKEN_LINE3="$(make -s -C dev-ops/local-slurm token3 | grep -E '^SLURM_JWT=' | tail -n1)"
if [[ -z "${TOKEN_LINE3}" ]]; then
    echo "ERROR: 'make token3' did not produce a SLURM_JWT=... line" >&2
    exit 1
fi

export TEST_SLURM_TOKEN3="${TOKEN_LINE3#SLURM_JWT=}"
echo "==> TEST_SLURM_TOKEN3 set (${#TEST_SLURM_TOKEN3} chars)"

TOKEN_LINE4="$(make -s -C dev-ops/local-slurm token4 | grep -E '^SLURM_JWT=' | tail -n1)"
if [[ -z "${TOKEN_LINE4}" ]]; then
    echo "ERROR: 'make token4' did not produce a SLURM_JWT=... line" >&2
    exit 1
fi

export TEST_SLURM_TOKEN4="${TOKEN_LINE4#SLURM_JWT=}"
echo "==> TEST_SLURM_TOKEN4 set (${#TEST_SLURM_TOKEN4} chars)"

export TEST_SLURM_API="http://localhost:6820"
export TEST_SLURM_USER="root"
export TEST_SLURM_API_VERSION="41"

# go test -tags integration -v -count=1 \
#    ./connectors/SLURM/Rest-Client/pkg/client/...

 go test -tags integration -v -count=1 \
    ./connectors/SLURM/Usage-Monitor/internal/smonitor/smonitor.go \
    ./connectors/SLURM/Usage-Monitor/internal/smonitor/smonitor_integration_test.go


#go test -tags integration -v -count=1 \
#  ./connectors/SLURM/Rest-Client/pkg/client/accounts.go \
#  ./connectors/SLURM/Rest-Client/pkg/client/associations.go \
#  ./connectors/SLURM/Rest-Client/pkg/client/client.go \
#  ./connectors/SLURM/Rest-Client/pkg/client/tres.go \
#  ./connectors/SLURM/Rest-Client/pkg/client/types.go \
#  ./connectors/SLURM/Rest-Client/pkg/client/integration_common.go \
#  ./connectors/SLURM/Rest-Client/pkg/client/associations_integration_test.go


#make -s -C dev-ops/local-slurm down