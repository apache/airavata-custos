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

set -euo pipefail

# Copy slurmdbd.conf with the 0600 perms that slurmdbd requires
# (slurm user's primary group is root/gid 0 — there is no 'slurm' group)
install -m 0600 -o slurm -g 0 /etc/slurm.readonly/slurmdbd.conf /etc/slurm/slurmdbd.conf
ln -sf /keys/jwt.hs256.key /etc/slurm/jwt.hs256.key

# Start munge. /var/run/munge must be world-readable (0755) so non-munge
# users (slurm) can open the munge socket; /var/log and /var/lib stay 0700.
install -d -m 0755 -o munge -g munge /var/run/munge
install -d -m 0700 -o munge -g munge /var/log/munge /var/lib/munge
runuser -u munge -- /usr/sbin/munged --force

# Wait for MariaDB to accept TCP connections. The compose healthcheck on the
# mariadb service already gates startup via depends_on, but this adds a
# belt-and-suspenders TCP probe (base image has no mariadb client binary).
until (exec 3<>/dev/tcp/mariadb/3306) 2>/dev/null; do
  echo "[slurmdbd] waiting for mariadb..."
  sleep 2
done
exec 3<&- 3>&- || true

exec slurmdbd -D -vvv
