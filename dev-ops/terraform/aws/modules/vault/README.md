<!--
    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.
-->

# Vault module

The module runs Vault on an EC2 Auto Scaling group with Raft storage and KMS auto-unseal, behind an Application Load
Balancer. Each node reads its TLS material from one Secrets Manager secret at boot. This page prepares that secret.

## TLS material

The secret is a JSON object with three keys:

| Key | Contents |
|---|---|
| `vault_cert` | The node certificate, PEM |
| `vault_ca` | The CA certificate, PEM. Same as `vault_cert` if self-signed. |
| `vault_pk` | The private key, PEM |

For development, generate a self-signed certificate from `resources/openssl-vault.cnf`:

```bash
cd dev-ops/terraform/aws/modules/vault/resources
openssl req -x509 -nodes -newkey rsa:2048 -keyout vault.key -out vault.crt -days 365 \
  -config openssl-vault.cnf -extensions req_ext
```

Create a Secret using AWS Secrets Manager and copy the `vault.crt` content into the `vault_cert` and `vault_ca`. The `vault.key` content should go to the key `vault_pk`
Then use the corresponding ARN value for variable `vault_secrets_manager_arn` in `terraform.tfvars`
