#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements. See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership. The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License. You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied. See the License for the
# specific language governing permissions and limitations
# under the License.

# End-to-end check of the signer compose stack:
# sign -> inspect cert -> SSH into the login container -> list -> revoke,
# plus the negative paths (no token, reserved username, expired cert, revoked serial).
#
# Run from dev-ops/compose after:
#   docker compose up -d db openbao openbao-init signer sshd-login

# no -e: failed checks are recorded and reported, not fatal
set -uo pipefail

SIGNER_URL="${SIGNER_URL:-http://localhost:8084}"
SERVICE_TOKEN="${SIGNER_SERVICE_TOKEN:-dev-service-token-not-for-production}"
SSH_PORT="${SSH_PORT:-2222}"
PRINCIPAL="${PRINCIPAL:-testuser}"
USER_EMAIL="testuser@example.edu"
USER_SUB="http://idp.example.edu/users/testuser"
REQUESTED_BY="app/e2e"

WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT

PASS=0
FAIL=0
check() {
  local name="$1" ok="$2"
  if [ "$ok" = "0" ]; then
    echo "PASS: $name"; PASS=$((PASS+1))
  else
    echo "FAIL: $name"; FAIL=$((FAIL+1))
  fi
}

json_get() {
  python3 -c "import sys,json; print(json.load(sys.stdin).get('$1',''))"
}

# --- sign
ssh-keygen -t ed25519 -N "" -q -f "$WORK_DIR/key"
PUBKEY=$(cat "$WORK_DIR/key.pub")

sign_body() {
  echo "{\"principal\":\"$1\",\"ttl_seconds\":$2,\"public_key\":\"$PUBKEY\",\"user_email\":\"$USER_EMAIL\",\"user_sub\":\"$USER_SUB\",\"requested_by\":\"$REQUESTED_BY\"${3:+,\"source_address\":\"$3\"}}"
}

sign() {
  curl -s -o "$WORK_DIR/resp.json" -w "%{http_code}" "$SIGNER_URL/api/v1/sign" \
    -H "Authorization: Bearer $SERVICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$(sign_body "$1" "$2" "${3:-}")"
}

code=$(curl -s -o /dev/null -w "%{http_code}" "$SIGNER_URL/api/v1/sign" \
  -H "Content-Type: application/json" -d "$(sign_body "$PRINCIPAL" 300)")
[ "$code" = "401" ]
check "sign without the service token rejected (HTTP $code)" $?

code=$(sign "$PRINCIPAL" 300)
[ "$code" = "200" ]
check "sign certificate (HTTP $code)" $?

write_cert() {
  json_get certificate < "$WORK_DIR/resp.json" > "$WORK_DIR/key-cert.pub"
}
write_cert
SERIAL=$(json_get serial_number < "$WORK_DIR/resp.json")

ssh-keygen -Lf "$WORK_DIR/key-cert.pub" > "$WORK_DIR/cert.txt"

grep -q "Key ID: \"$USER_EMAIL:$REQUESTED_BY\"" "$WORK_DIR/cert.txt"
check "key ID names the user and who asked" $?

grep -q "user-sub@custos.apache.org" "$WORK_DIR/cert.txt" && \
  grep -q "user-email@custos.apache.org" "$WORK_DIR/cert.txt" && \
  grep -q "requested-by@custos.apache.org" "$WORK_DIR/cert.txt"
check "certificate carries the user and requester extensions" $?

SSH_OPTS=(-p "$SSH_PORT" -i "$WORK_DIR/key" -o "CertificateFile=$WORK_DIR/key-cert.pub"
  -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null
  -o PasswordAuthentication=no -o ConnectTimeout=5 -o LogLevel=ERROR)

out=$(ssh "${SSH_OPTS[@]}" "$PRINCIPAL@127.0.0.1" whoami)
[ "$out" = "$PRINCIPAL" ]
check "SSH login as $PRINCIPAL" $?

# --- audit row
count=$(docker compose exec -T db psql -U admin -d custos_signer -tA \
  -c "SELECT COUNT(*) FROM certificate_issuance_logs WHERE serial_number = $SERIAL AND user_sub = '$USER_SUB' AND user_email = '$USER_EMAIL' AND requested_by = '$REQUESTED_BY'")
[ "$count" = "1" ]
check "audit log row for serial $SERIAL names the user and who asked" $?

# --- a list filtered to the user finds it, and a list for someone else does not
listed() {
  curl -s -G "$SIGNER_URL/api/v1/certificates" -H "Authorization: Bearer $SERVICE_TOKEN" \
    --data-urlencode "user_sub=$1" | python3 -c "import sys,json; print(any(c['serial_number'] == $SERIAL for c in json.load(sys.stdin)['certificates']))"
}
[ "$(listed "$USER_SUB")" = "True" ] && [ "$(listed "someone-else")" = "False" ]
check "list filter returns only that user's certificates" $?

# --- reserved username
code=$(sign "root" 300)
[ "$code" = "403" ]
check "reserved username rejected (HTTP $code)" $?

# --- revoke through the API: recorded with who did it, and shown on the certificate
code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$SIGNER_URL/api/v1/certificates/$SERIAL/revoke" \
  -H "Authorization: Bearer $SERVICE_TOKEN" -H "Content-Type: application/json" \
  -d '{"reason":"","revoked_by":"self"}')
[ "$code" = "200" ]
check "revoke with an empty reason (HTTP $code)" $?

revoked=$(curl -s "$SIGNER_URL/api/v1/certificates/$SERIAL" -H "Authorization: Bearer $SERVICE_TOKEN" | \
  python3 -c "import sys,json; c=json.load(sys.stdin); print(c['revoked'], c['revoked_by'])")
[ "$revoked" = "True self" ]
check "certificate shows revoked by self (saw: $revoked)" $?

# --- source address: sshd accepts the certificate only from the listed addresses
code=$(sign "$PRINCIPAL" 300 "203.0.113.5,2001:db8::/32,0.0.0.0/0")
[ "$code" = "200" ]
check "sign certificate with a source address list (HTTP $code)" $?
write_cert
out=$(ssh "${SSH_OPTS[@]}" "$PRINCIPAL@127.0.0.1" whoami)
[ "$out" = "$PRINCIPAL" ]
check "SSH login from an address on the list" $?

code=$(sign "$PRINCIPAL" 300 "192.0.2.0/24")
[ "$code" = "200" ]
check "sign certificate limited to another network (HTTP $code)" $?
write_cert
RESTRICTED_SERIAL=$(json_get serial_number < "$WORK_DIR/resp.json")
ssh-keygen -Lf "$WORK_DIR/key-cert.pub" | grep -q "source-address 192.0.2.0/24"
check "certificate carries the source address" $?
if ssh "${SSH_OPTS[@]}" "$PRINCIPAL@127.0.0.1" true 2>/dev/null; then
  check "SSH login from an address not on the list rejected" 1
else
  check "SSH login from an address not on the list rejected" 0
fi

stored=$(curl -s "$SIGNER_URL/api/v1/certificates/$RESTRICTED_SERIAL" -H "Authorization: Bearer $SERVICE_TOKEN" | json_get source_address)
[ "$stored" = "192.0.2.0/24" ]
check "certificate record shows the source address (saw: $stored)" $?

code=$(sign "$PRINCIPAL" 300 "10.0.0.5/24")
[ "$code" = "400" ]
check "CIDR with host bits set rejected (HTTP $code)" $?

# --- expired cert: TTL 1s, wait, connection must fail
code=$(sign "$PRINCIPAL" 1)
[ "$code" = "200" ]
check "sign short-lived certificate (HTTP $code)" $?
write_cert
sleep 2
if ssh "${SSH_OPTS[@]}" "$PRINCIPAL@127.0.0.1" true 2>/dev/null; then
  check "expired certificate rejected" 1
else
  check "expired certificate rejected" 0
fi

# --- revoked serial: sign, revoke via KRL on the login node, must fail, then reset
code=$(sign "$PRINCIPAL" 300)
[ "$code" = "200" ]
check "sign certificate to revoke (HTTP $code)" $?
write_cert

docker compose exec -T sshd-login sh -c 'cat > /tmp/revoke-cert.pub && ssh-keygen -kuf /etc/ssh/revoked_keys.krl /tmp/revoke-cert.pub' < "$WORK_DIR/key-cert.pub"
if ssh "${SSH_OPTS[@]}" "$PRINCIPAL@127.0.0.1" true 2>/dev/null; then
  check "revoked certificate rejected" 1
else
  check "revoked certificate rejected" 0
fi
docker compose exec -T sshd-login ssh-keygen -kf /etc/ssh/revoked_keys.krl

# --- serial races: the signer detects them by matching OpenBao's CAS
# message, so pin the mapping here; a reword upstream would surface as a
# 500 instead of a 409. Passes when no race occurs.
codes=$(for _ in $(seq 1 20); do
  curl -s -o /dev/null -w "%{http_code}\n" "$SIGNER_URL/api/v1/sign" \
    -H "Authorization: Bearer $SERVICE_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$(sign_body "$PRINCIPAL" 300)" &
done | sort -u | tr '\n' ' ')
if [ -n "$(echo "$codes" | tr ' ' '\n' | grep -v -E '^(200|409)?$')" ] || ! echo "$codes" | grep -qw 200; then
  check "concurrent serial conflicts return 409, not 500 (saw: $codes)" 1
else
  check "concurrent serial conflicts return 409, not 500 (saw: $codes)" 0
fi

echo
echo "$PASS passed, $FAIL failed"
[ "$FAIL" = "0" ]
