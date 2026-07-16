#!/usr/bin/env bash
# PEARC26 live setup: creates the tutorial project, allocation, and resource
# grant THROUGH THE API so the provisioning events fire (SLURM account +
# grant limits appear on the cluster), then points the PEARC26 access event
# at the new allocation. Run AFTER seeds 01-03 and with the backend running.
#
# Auth: CILogon device flow as the admin (client credentials from .env);
# export ACCESS_TOKEN to skip the device flow with a ready bearer.
#
# Usage: dev-ops/compose/seeds/pearc26/setup-live.sh

set -euo pipefail
cd "$(dirname "$0")/../../../.."

API="${CUSTOS_API:-http://localhost:8080}"
DB_EXEC=(docker exec -i custos_db mariadb -u admin -padmin custos)

set -a; source .env; set +a
: "${CUSTOS_CLUSTER_ID:?missing in .env}"

existing=$("${DB_EXEC[@]}" -N -e "SELECT COUNT(*) FROM compute_allocations WHERE name='pearc26-tutorial' AND status='ACTIVE'" 2>/dev/null || echo 0)
if [ "$existing" != "0" ]; then
  echo "An ACTIVE 'pearc26-tutorial' allocation already exists; tear it down first (or repoint access_events manually)."
  exit 1
fi

if [ -z "${ACCESS_TOKEN:-}" ]; then
  : "${OIDC_CLIENT_ID:?missing in .env}" "${OIDC_CLIENT_SECRET:?missing in .env}"
  ISSUER="${OIDC_ISSUER_URL%/}"
  echo "== CILogon device authorization ($ISSUER)"
  dev=$(curl -sf -X POST "$ISSUER/oauth2/device_authorization" \
    -d "client_id=$OIDC_CLIENT_ID" -d "client_secret=$OIDC_CLIENT_SECRET" \
    -d "scope=openid profile email org.cilogon.userinfo")
  device_code=$(echo "$dev" | python3 -c "import json,sys; print(json.load(sys.stdin)['device_code'])")
  echo "$dev" | python3 -c "import json,sys; d=json.load(sys.stdin); print('Authenticate at:', d.get('verification_uri_complete') or d.get('verification_uri')); print('Code:', d.get('user_code',''))"
  interval=$(echo "$dev" | python3 -c "import json,sys; print(json.load(sys.stdin).get('interval',5))")
  echo "Waiting for authentication..."
  while true; do
    sleep "$interval"
    tok=$(curl -s -X POST "$ISSUER/oauth2/token" \
      -d "client_id=$OIDC_CLIENT_ID" -d "client_secret=$OIDC_CLIENT_SECRET" \
      -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
      -d "device_code=$device_code")
    if echo "$tok" | grep -q access_token; then break; fi
    if echo "$tok" | grep -qE "expired_token|access_denied"; then
      echo "Device authorization failed: $tok"; exit 1
    fi
  done
  # The backend verifies the id_token (the access token is opaque or
  # carries a different audience). Print redacted claims for diagnosis.
  ACCESS_TOKEN=$(echo "$tok" | python3 -c "
import json,sys
d=json.load(sys.stdin)
print(d.get('id_token') or d.get('access_token',''))")
  echo "$tok" | python3 -c "
import base64, json, sys
d = json.load(sys.stdin)
def peek(name):
    t = d.get(name) or ''
    if t.count('.') != 2:
        print(f'  {name}: opaque/absent')
        return
    pay = t.split('.')[1]
    pay += '=' * (-len(pay) % 4)
    c = json.loads(base64.urlsafe_b64decode(pay))
    print(f'  {name}: iss={c.get(\"iss\")} aud={c.get(\"aud\")} sub={str(c.get(\"sub\"))[:40]}')
peek('access_token'); peek('id_token')
print('  using:', 'id_token' if d.get('id_token') else 'access_token')"
  echo "Authenticated."
fi

capi() { # method path [json-body] -- prints the body, aborts loudly on non-2xx
  local method="$1" path="$2" body="${3:-}" out code
  out=$(mktemp)
  if [ -n "$body" ]; then
    code=$(curl -s -o "$out" -w "%{http_code}" -X "$method" -H "Authorization: Bearer $ACCESS_TOKEN" -H "Content-Type: application/json" "$API$path" -d "$body")
  else
    code=$(curl -s -o "$out" -w "%{http_code}" -X "$method" -H "Authorization: Bearer $ACCESS_TOKEN" "$API$path")
  fi
  if [ "${code:0:1}" != "2" ]; then
    echo "" >&2
    echo "FAILED: $method $path -> HTTP $code" >&2
    cat "$out" >&2; echo "" >&2
    rm -f "$out"
    exit 1
  fi
  cat "$out"; rm -f "$out"
}

echo "== Preflight: who am I against the backend"
me=$(capi GET /me)
ADMIN_ID=$(echo "$me" | python3 -c "
import json,sys
d=json.load(sys.stdin)
u=d.get('user') or {}
print('  user:', u.get('email'), '(', u.get('id'), ')', file=sys.stderr)
print('  privileges:', len(d.get('privileges') or []), file=sys.stderr)
print(u.get('id') or '')")
[ -n "$ADMIN_ID" ] || { echo "Could not resolve the caller from /me: $me"; exit 1; }
if ! echo "$me" | grep -q "projects:write"; then
  echo "This identity lacks core:projects:write. Authenticate as the ADMIN"
  echo "identity at the CILogon prompt (the linked admin email), then re-run."
  exit 1
fi

echo "== Creating project"
# The authenticated admin is the PI: the seed approver has no login.
project=$(capi POST /projects "{\"originated_id\":\"PEARC26\",\"title\":\"PEARC26 Tutorial Workshop\",\"origination\":\"pearc26\",\"project_pi_id\":\"$ADMIN_ID\",\"status\":\"ACTIVE\"}")
project_id=$(echo "$project" | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")
echo "  project $project_id"

echo "== Creating allocation (fires the SLURM account event)"
start=$(date -u +%Y-%m-%dT%H:%M:%SZ)
end=$(date -u -v+45d +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || date -u -d "+45 days" +%Y-%m-%dT%H:%M:%SZ)
alloc=$(capi POST /compute-allocations "{\"project_id\":\"$project_id\",\"name\":\"pearc26-tutorial\",\"status\":\"ACTIVE\",\"compute_cluster_id\":\"$CUSTOS_CLUSTER_ID\",\"initial_su_amount\":25000,\"start_time\":\"$start\",\"end_time\":\"$end\"}")
alloc_id=$(echo "$alloc" | python3 -c "import json,sys; print(json.load(sys.stdin)['id'])")
echo "  allocation $alloc_id ($start -> $end)"

echo "== Attaching the resource grant (fires the SLURM limits event)"
capi POST "/compute-allocations/$alloc_id/resources" '{"compute_allocation_resource_id":"pearc26-res-debug-cpu","resource_amount":4,"resource_time":87840}' >/dev/null
echo "  grant: cpu x4 for 87840 minutes on partition debug"

echo "== Recording PI membership and the PEARC26 access event"
"${DB_EXEC[@]}" <<SQL
INSERT IGNORE INTO project_memberships (project_id, user_id, role, added_time)
VALUES ('$project_id', '$ADMIN_ID', 'PI', NOW(6));
INSERT INTO access_events (code, compute_allocation_id, organization_id)
VALUES ('PEARC26', '$alloc_id', 'pearc26-org')
ON DUPLICATE KEY UPDATE compute_allocation_id = '$alloc_id';
SQL

echo
echo "Done. Verify on the cluster (allow a few seconds for the events):"
echo "  sacctmgr -n show account pearc26-tutorial"
echo "  sacctmgr -n show assoc account=pearc26-tutorial format=Account,User,Partition,GrpTRES,GrpTRESMins"
