#!/usr/bin/env bash
set -euo pipefail

install -m 0644 /etc/slurm.readonly/slurm.conf /etc/slurm/slurm.conf
install -m 0644 /etc/slurm.readonly/cgroup.conf /etc/slurm/cgroup.conf
ln -sf /keys/jwt.hs256.key /etc/slurm/jwt.hs256.key

# Ensure StateSaveLocation exists (the slurmctld-state named volume is empty
# on first boot; /var/spool/slurm itself is created by the base image).
install -d -m 0755 -o slurm -g 0 /var/spool/slurm/ctld

install -d -m 0755 -o munge -g munge /var/run/munge
install -d -m 0700 -o munge -g munge /var/log/munge /var/lib/munge
runuser -u munge -- /usr/sbin/munged --force

# Bootstrap accounting in the background after slurmctld comes up
( sleep 5; /usr/local/bin/bootstrap-accounting.sh ) &

exec slurmctld -D -vvv
