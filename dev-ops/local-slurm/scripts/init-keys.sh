#!/usr/bin/env bash
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
