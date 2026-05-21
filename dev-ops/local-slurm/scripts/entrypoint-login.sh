#!/usr/bin/env bash
set -euo pipefail

install -m 0644 /etc/slurm.readonly/slurm.conf /etc/slurm/slurm.conf
ln -sf /keys/jwt.hs256.key /etc/slurm/jwt.hs256.key

install -d -m 0755 -o munge -g munge /var/run/munge
install -d -m 0700 -o munge -g munge /var/log/munge /var/lib/munge
runuser -u munge -- /usr/sbin/munged --force

# Ensure the default test user exists
id -u testuser >/dev/null 2>&1 || useradd -m -s /bin/bash testuser
id -u testuser2 >/dev/null 2>&1 || useradd -m -s /bin/bash testuser2
id -u testuser3 >/dev/null 2>&1 || useradd -m -s /bin/bash testuser3

# Start sshd in the foreground
exec /usr/sbin/sshd -D -e
