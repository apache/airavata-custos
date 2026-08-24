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
# token -> sign -> inspect cert -> SSH into the login container,
# plus the negative paths (wrong principal, expired cert, revoked serial).
#
# Run from dev-ops/compose after:
#   docker compose up -d db keycloak openldap openbao openbao-init \
#                        signer signer-seed sshd-login

# no -e: failed checks are recorded and reported, not fatal
set -uo pipefail

SIGNER_URL="${SIGNER_URL:-http://localhost:8084}"
CLIENT_ID="${CLIENT_ID:-custos:bench}"
CLIENT_SECRET="${CLIENT_SECRET:-bench-secret}"
SSH_PORT="${SSH_PORT:-2222}"
PRINCIPAL="${PRINCIPAL:-testuser}"

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

# --- token (fetched inside the compose network so the issuer URL is
# --- resolvable by the signer when it fetches the JWKS)
TOKEN=$(docker compose exec -T sshd-login curl -sf \
  "http://keycloak:8080/realms/custos/protocol/openid-connect/token" \
  -d grant_type=password -d client_id=signer-cli \
  -d username=testuser -d password=testuser -d scope=openid | json_get access_token)
[ -n "$TOKEN" ]
check "obtain user token" $?

# --- sign
ssh-keygen -t ed25519 -N "" -q -f "$WORK_DIR/key"
PUBKEY=$(cat "$WORK_DIR/key.pub")

sign() {
  local principal="$1" ttl="$2"
  curl -s -o "$WORK_DIR/resp.json" -w "%{http_code}" "$SIGNER_URL/api/v1/sign" \
    -H "X-Client-Id: $CLIENT_ID" -H "X-Client-Secret: $CLIENT_SECRET" \
    -H "Content-Type: application/json" \
    -d "{\"principal\":\"$principal\",\"ttl_seconds\":$ttl,\"public_key\":\"$PUBKEY\",\"user_access_token\":\"$TOKEN\"}"
}

code=$(sign "$PRINCIPAL" 300)
[ "$code" = "200" ]
check "sign certificate (HTTP $code)" $?

write_cert() {
  # response carries the raw cert blob; the type prefix completes the OpenSSH line
  echo "ssh-ed25519-cert-v01@openssh.com $(json_get certificate < "$WORK_DIR/resp.json")" > "$WORK_DIR/key-cert.pub"
}
write_cert
SERIAL=$(json_get serial_number < "$WORK_DIR/resp.json")

ssh-keygen -Lf "$WORK_DIR/key-cert.pub" | grep -q "$PRINCIPAL"
check "certificate carries principal" $?

SSH_OPTS=(-p "$SSH_PORT" -i "$WORK_DIR/key" -o "CertificateFile=$WORK_DIR/key-cert.pub"
  -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null
  -o PasswordAuthentication=no -o ConnectTimeout=5 -o LogLevel=ERROR)

out=$(ssh "${SSH_OPTS[@]}" "$PRINCIPAL@127.0.0.1" whoami)
[ "$out" = "$PRINCIPAL" ]
check "SSH login as $PRINCIPAL" $?

# --- audit row
count=$(docker compose exec -T db mariadb -u admin -padmin custos_signer -N \
  -e "SELECT COUNT(*) FROM certificate_issuance_logs WHERE serial_number = $SERIAL")
[ "$count" = "1" ]
check "audit log row for serial $SERIAL" $?

# --- wrong principal: subject resolves to testuser, so testuser2 must be denied
code=$(sign "testuser2" 300)
[ "$code" = "403" ]
check "wrong principal rejected (HTTP $code)" $?

# --- expired cert: TTL 1s, wait, connection must fail
code=$(sign "$PRINCIPAL" 1)
[ "$code" = "200" ]
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
    -H "X-Client-Id: $CLIENT_ID" -H "X-Client-Secret: $CLIENT_SECRET" \
    -H "Content-Type: application/json" \
    -d "{\"principal\":\"$PRINCIPAL\",\"ttl_seconds\":300,\"public_key\":\"$PUBKEY\",\"user_access_token\":\"$TOKEN\"}" &
done | sort -u | tr '\n' ' ')
if echo "$codes" | grep -qw 500; then
  check "concurrent serial conflicts return 409, not 500 (saw: $codes)" 1
else
  check "concurrent serial conflicts return 409, not 500 (saw: $codes)" 0
fi

echo
echo "$PASS passed, $FAIL failed"
[ "$FAIL" = "0" ]
