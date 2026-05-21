#!/usr/bin/env bash
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
