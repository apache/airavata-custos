#!/bin/sh

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

# One-shot dev bootstrap: init + unseal the store, enable the mounts the
# services expect, create a fixed dev token, and seed the signer's
# per-client validation credentials. Idempotent across restarts.

set -eu

DEV_TOKEN="00000000-0000-0000-0000-000000000000"
KEYS_FILE="/keys/init.txt"

echo "waiting for openbao..."
until bao status >/dev/null 2>&1 || [ "$?" = "2" ]; do
  sleep 2
done

st=$(bao status 2>/dev/null || true)

if echo "$st" | grep -q "Initialized.*false"; then
  echo "initializing..."
  bao operator init -key-shares=1 -key-threshold=1 > "$KEYS_FILE"
fi

UNSEAL_KEY=$(grep "Unseal Key 1:" "$KEYS_FILE" | awk '{print $NF}')
ROOT_TOKEN=$(grep "Initial Root Token:" "$KEYS_FILE" | awk '{print $NF}')

st=$(bao status 2>/dev/null || true)
if echo "$st" | grep -q "Sealed.*true"; then
  echo "unsealing..."
  bao operator unseal "$UNSEAL_KEY" >/dev/null
fi

export BAO_TOKEN="$ROOT_TOKEN"

mounts=$(bao secrets list)
echo "$mounts" | grep -q "^ssh-ca/" || bao secrets enable -path=ssh-ca kv-v2
echo "$mounts" | grep -q "^secret/" || bao secrets enable -version=1 -path=secret kv
echo "$mounts" | grep -q "^resourcesecret/" || bao secrets enable -version=1 -path=resourcesecret kv

# fixed token so service configs don't depend on the generated root token
bao token lookup "$DEV_TOKEN" >/dev/null 2>&1 || \
  bao token create -id="$DEV_TOKEN" -policy=root -orphan >/dev/null

# principal validation credentials for the seeded signer client
bao kv put ssh-ca/custos/bench/validation \
  type=ldap \
  ldap_url=ldap://openldap:389 \
  bind_dn="cn=admin,dc=custos,dc=example" \
  bind_password=admin \
  base_dn="ou=people,dc=custos,dc=example" \
  search_filter="(&(objectClass=posixAccount)(voPersonExternalID=%s))" \
  username_attribute=uid >/dev/null

echo "openbao bootstrap complete"
