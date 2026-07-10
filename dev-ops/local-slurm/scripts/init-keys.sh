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

# scripts/init-keys.sh — generate munge + JWT keys into shared volumes if missing.
set -euo pipefail

MUNGE_KEY=/etc/munge/munge.key
# jwt-key volume is mounted at /keys; each service symlinks into /etc/slurm/
JWT_KEY=/keys/jwt.hs256.key

if [[ ! -s "$MUNGE_KEY" ]]; then
  echo "[init-keys] generating $MUNGE_KEY"
  install -d -m 0700 -o munge -g munge /etc/munge
  dd if=/dev/urandom of="$MUNGE_KEY" bs=1 count=1024 status=none
  chown munge:munge "$MUNGE_KEY"
  chmod 0400 "$MUNGE_KEY"
else
  echo "[init-keys] $MUNGE_KEY already present"
fi

if [[ ! -s "$JWT_KEY" ]]; then
  echo "[init-keys] generating $JWT_KEY"
  install -d -m 0755 /keys
  openssl rand -base64 32 | tr -d '\n' > "$JWT_KEY"
  # slurm user's primary group is root (gid 0) in the base image; chown accordingly
  chown slurm:0 "$JWT_KEY"
  chmod 0400 "$JWT_KEY"
else
  echo "[init-keys] $JWT_KEY already present"
fi

echo "[init-keys] done"
