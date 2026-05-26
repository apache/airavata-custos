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

# slurm.conf sets AccountingStorageEnforce=...,qos,..., which means every
# association must have at least one allowed QoS and a default QoS or the
# controller rejects submissions with the misleading error "Invalid account
# or account/partition combination specified". Ensure the built-in `normal`
# QoS is allowed on the cluster and set as the default for new associations.
if ! sacctmgr -in show qos format=name | grep -qw "normal"; then
  sacctmgr -i add qos normal
fi
sacctmgr -i modify cluster "$CLUSTER" set QOS=normal DefaultQOS=normal >/dev/null

if ! sacctmgr -in show account format=account | grep -qw "root"; then
  sacctmgr -i add account root Description="root account" Organization="$CLUSTER"
fi
if ! sacctmgr -in show user format=user | grep -qw "root"; then
  sacctmgr -i add user root Account=root AdminLevel=Administrator
fi

touch "$SENTINEL"
echo "[bootstrap] done"
