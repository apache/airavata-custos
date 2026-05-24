#!/usr/bin/env bash
#
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

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

cleanup() {
  make -s -C dev-ops/local-amie down >/dev/null 2>&1 || true
}
trap cleanup EXIT

make -s -C dev-ops/local-amie down >/dev/null 2>&1 || true
make -s -C dev-ops/local-amie up

export DATABASE_DSN="admin:admin@tcp(localhost:3307)/custos?parseTime=true&charset=utf8mb4&multiStatements=true"
export AMIE_BASE_URL="http://localhost:8181"
export AMIE_SITE_CODE="TESTSITE"
export AMIE_API_KEY="dev"
export AMIE_CLUSTER_ID="00000000-0000-0000-0000-000000000001"
export TEST_AMIE_INFRA_READY=1

go test -tags integration -v -count=1 -p 1 -timeout 10m ./connectors/ACCESS/AMIE-Processor/...
