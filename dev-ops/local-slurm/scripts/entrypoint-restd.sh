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

install -m 0644 /etc/slurm.readonly/slurm.conf /etc/slurm/slurm.conf
ln -sf /keys/jwt.hs256.key /etc/slurm/jwt.hs256.key

install -d -m 0755 -o munge -g munge /var/run/munge
install -d -m 0700 -o munge -g munge /var/log/munge /var/lib/munge
runuser -u munge -- /usr/sbin/munged --force

# slurmrestd must not run as root.
# SLURM_JWT=daemon makes slurmrestd trust its own internal JWT for daemon-to-daemon calls;
# external requests still require X-SLURM-USER-TOKEN.
# SLURMRESTD_SECURITY flags:
#   disable_unshare_sysv/files: Docker denies CLONE_NEWIPC without CAP_SYS_ADMIN,
#     which we don't want to grant; skip those hardening steps.
#   disable_user_check: the base image's 'slurm' user is slurm:0 (no slurm group
#     exists), matching how slurmdbd/slurmctld already run. slurmrestd's default
#     check rejects root primary group; we opt out since the daemon itself is not
#     running as uid 0.
exec runuser -u slurm -- env \
  SLURM_JWT=daemon \
  SLURMRESTD_SECURITY=disable_unshare_sysv,disable_unshare_files,disable_user_check \
  slurmrestd -f /etc/slurm/slurm.conf -a rest_auth/jwt 0.0.0.0:6820 -vvv
