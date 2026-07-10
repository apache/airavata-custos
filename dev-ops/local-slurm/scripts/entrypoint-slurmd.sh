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
install -m 0644 /etc/slurm.readonly/cgroup.conf /etc/slurm/cgroup.conf
install -m 0644 /etc/slurm.readonly/gres.conf /etc/slurm/gres.conf
ln -sf /keys/jwt.hs256.key /etc/slurm/jwt.hs256.key

# Ensure SlurmdSpoolDir exists (slurm.conf sets it to /var/spool/slurm/d;
# /var/spool/slurm itself is created by the base image but the subdir is not).
install -d -m 0755 -o slurm -g 0 /var/spool/slurm/d

# Create two fake GPU device files so slurmd can register Gres=gpu:2 against
# distinct File= entries. These are just /dev/null-style sinks — there are no
# real GPUs. gres.conf references /dev/nullgpu0 and /dev/nullgpu1.
for i in 0 1; do
  [ -e "/dev/nullgpu${i}" ] || mknod -m 0666 "/dev/nullgpu${i}" c 1 3
done

install -d -m 0755 -o munge -g munge /var/run/munge
install -d -m 0700 -o munge -g munge /var/log/munge /var/lib/munge
runuser -u munge -- /usr/sbin/munged --force

exec slurmd -D -N "${SLURMD_NODENAME:-$(hostname)}" -vvv
