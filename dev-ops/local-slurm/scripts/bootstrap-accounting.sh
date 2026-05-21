#!/usr/bin/env bash
# Idempotent: creates the demo cluster, root account, root admin user.
set -euo pipefail

SENTINEL=/var/spool/slurm/ctld/.bootstrap-done

if [[ -f "$SENTINEL" ]]; then
  echo "[bootstrap] sentinel present, skipping"
  exit 0
fi

# Wait until slurmdbd answers
until sacctmgr -i show cluster >/dev/null 2>&1; do
  echo "[bootstrap] waiting for slurmdbd..."
  sleep 2
done

CLUSTER="${CLUSTER_NAME:-artisan}"
if ! sacctmgr -in show cluster format=cluster | grep -qw "$CLUSTER"; then
  sacctmgr -i add cluster "$CLUSTER"
fi
if ! sacctmgr -in show account format=account | grep -qw "root"; then
  sacctmgr -i add account root Description="root account" Organization="$CLUSTER"
fi
if ! sacctmgr -in show user format=user | grep -qw "root"; then
  sacctmgr -i add user root Account=root AdminLevel=Administrator
fi

touch "$SENTINEL"
echo "[bootstrap] done"
