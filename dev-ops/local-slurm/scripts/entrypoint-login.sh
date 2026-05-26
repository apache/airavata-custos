#!/usr/bin/env bash
set -euo pipefail

install -m 0644 /etc/slurm.readonly/slurm.conf /etc/slurm/slurm.conf
ln -sf /keys/jwt.hs256.key /etc/slurm/jwt.hs256.key

install -d -m 0755 -o munge -g munge /var/run/munge
install -d -m 0700 -o munge -g munge /var/log/munge /var/lib/munge
runuser -u munge -- /usr/sbin/munged --force

# testuser{,2,3} are baked into the base image with fixed UIDs so they
# match across login/slurmctld/slurmd. Nothing to create here.

# Start sshd in the foreground
exec /usr/sbin/sshd -D -e
