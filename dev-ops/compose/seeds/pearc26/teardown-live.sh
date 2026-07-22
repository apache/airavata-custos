#!/usr/bin/env bash
# Tear down the PEARC26 tutorial allocation so setup-live.sh can recreate it
# from scratch: removes the Slurm account (and every association under it) and
# the custos project/allocation/usage/event rows. Leaves the org, cluster,
# partition catalog, rate, and approver role (seeds 01-03) in place.
#
# Usage: dev-ops/compose/seeds/pearc26/teardown-live.sh [--yes]

set -euo pipefail
cd "$(dirname "$0")/../../../.."

ACCOUNT="pearc26-tutorial"
CONFIRM="${1:-}"

set -a; source .env; set +a

# DB credentials from DATABASE_DSN when present (deployment hosts differ
# from the local compose defaults).
DB_USER=admin; DB_PASS=admin
if grep -q "^DATABASE_DSN=" .env; then
  _dsn=$(grep "^DATABASE_DSN=" .env | cut -d= -f2- | tr -d '"' | tr -d "'")
  DB_USER=$(echo "$_dsn" | cut -d: -f1)
  DB_PASS=$(echo "$_dsn" | sed -E 's/^[^:]*:([^@]*)@.*/\1/')
fi
DB_EXEC=(docker exec -i custos_db mariadb -u"$DB_USER" -p"$DB_PASS" custos)

echo "== Current state"
alloc_ids=$("${DB_EXEC[@]}" -N -e "SELECT id FROM compute_allocations WHERE name='$ACCOUNT'" 2>/dev/null || true)
echo "  custos allocations named $ACCOUNT: ${alloc_ids:-none}"
if [ -n "${SLURM_API_URL:-}" ] && [ -n "${SLURM_TOKEN:-}" ]; then
  ver="v0.0.${SLURM_API_VERSION:-40}"
  n=$(curl -s -H "X-SLURM-USER-NAME: ${SLURM_API_USERNAME:-nexus-provisioner}" -H "X-SLURM-USER-TOKEN: $SLURM_TOKEN" \
    "$SLURM_API_URL/slurmdb/$ver/associations?account=$ACCOUNT" \
    | python3 -c "import json,sys; print(len(json.load(sys.stdin).get('associations',[])))" 2>/dev/null || echo "?")
  echo "  slurm associations under $ACCOUNT: $n"
else
  echo "  SLURM_API_URL/SLURM_TOKEN not set; Slurm teardown will be skipped"
fi

if [ -z "$alloc_ids" ] && { [ -z "${SLURM_API_URL:-}" ] || [ "${n:-0}" = "0" ]; }; then
  echo "Nothing to tear down."; exit 0
fi

if [ "$CONFIRM" != "--yes" ]; then
  read -r -p "Tear all of this down? [y/N] " ans
  [ "$ans" = "y" ] || { echo "Aborted."; exit 1; }
fi

if [ -n "${SLURM_API_URL:-}" ] && [ -n "${SLURM_TOKEN:-}" ]; then
  echo "== Removing the Slurm account (cascades its associations)"
  removed=$(curl -s -X DELETE \
    -H "X-SLURM-USER-NAME: ${SLURM_API_USERNAME:-nexus-provisioner}" -H "X-SLURM-USER-TOKEN: $SLURM_TOKEN" \
    "$SLURM_API_URL/slurmdb/$ver/account/$ACCOUNT" \
    | python3 -c "import json,sys; d=json.load(sys.stdin); print(len(d.get('removed_associations',[])), '/', (d.get('errors') or 'ok'))" 2>/dev/null || echo "delete failed")
  echo "  removed associations / errors: $removed"
fi

echo "== Cleaning custos allocation rows"
for aid in $alloc_ids; do
  "${DB_EXEC[@]}" <<SQL
DELETE FROM compute_allocation_usages WHERE compute_allocation_id='$aid';
DELETE FROM compute_allocation_membership_resource_overrides WHERE compute_allocation_membership_id IN
  (SELECT id FROM compute_allocation_memberships WHERE compute_allocation_id='$aid');
DELETE FROM compute_allocation_memberships WHERE compute_allocation_id='$aid';
DELETE FROM compute_allocation_resource_mappings WHERE compute_allocation_id='$aid';
-- Access requests reference the event by code, so clear them (and their event
-- log) before the access_events row they point at.
DELETE are FROM access_request_events are
  JOIN access_requests ar ON ar.id = are.access_request_id
  JOIN access_events ae ON ae.code = ar.event_code
 WHERE ae.compute_allocation_id='$aid';
DELETE ar FROM access_requests ar
  JOIN access_events ae ON ae.code = ar.event_code
 WHERE ae.compute_allocation_id='$aid';
DELETE FROM access_events WHERE compute_allocation_id='$aid';
DELETE FROM compute_allocations WHERE id='$aid';
SQL
  echo "  allocation $aid removed"
done

# Drop the tutorial project only when nothing else hangs off it.
"${DB_EXEC[@]}" <<SQL 2>/dev/null || true
DELETE pm FROM project_memberships pm
  JOIN projects p ON p.id = pm.project_id
 WHERE p.originated_id='PEARC26'
   AND NOT EXISTS (SELECT 1 FROM compute_allocations a WHERE a.project_id = p.id);
DELETE p FROM projects p
 WHERE p.originated_id='PEARC26'
   AND NOT EXISTS (SELECT 1 FROM compute_allocations a WHERE a.project_id = p.id);
SQL
echo "  tutorial project removed (if it had no other allocations)"

echo
echo "Done. Re-run setup-live.sh to recreate the allocation through the API."
