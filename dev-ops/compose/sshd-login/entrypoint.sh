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

set -e

# Fetching the CA public key also triggers CA creation on first run.
echo "fetching CA public key from ${SIGNER_URL}..."
until curl -sf \
    -H "X-Client-Id: ${CLIENT_ID}" \
    -H "X-Client-Secret: ${CLIENT_SECRET}" \
    "${SIGNER_URL}/api/v1/ca-public-key" \
    -o /etc/ssh/trusted-user-ca-keys.pub; do
  sleep 2
done
echo "CA key installed: $(cut -c1-40 /etc/ssh/trusted-user-ca-keys.pub)..."

[ -s /etc/ssh/revoked_keys.krl ] || ssh-keygen -kf /etc/ssh/revoked_keys.krl
ssh-keygen -A

exec /usr/sbin/sshd -D -e
