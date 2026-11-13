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

# Signer

The Signer connector puts the [SSH certificate signer](../../extensions/SSH-Certificate-Signer/README.md) behind the
Custos server. It registers the `signer:` privileges and forwards the certificate routes to the signer with the caller's
own token, so the signer applies its owner and privilege checks to the real user.

## Routes

| Route                                                  | Needs                       | Forwards to                                                       |
|--------------------------------------------------------|-----------------------------|-------------------------------------------------------------------|
| `GET /connectors/signer/certificates`                  | `signer:certificates:read`  | `GET /api/v1/admin/certificates`                                  |
| `GET /connectors/signer/certificates/{serial}`         | `signer:certificates:read`  | `GET /api/v1/admin/certificates/{serial}`                         |
| `POST /connectors/signer/certificates/{serial}/revoke` | `signer:certificates:write` | `POST /api/v1/certificates/{serial}/revoke`                       |
| `GET /connectors/signer/me/certificates`               | Signed in                   | `GET /api/v1/certificates`                                        |
| `GET /connectors/signer/me/certificates/{serial}`      | Signed in                   | `GET /api/v1/certificates/{serial}`                               |
| `GET /connectors/signer/status`                        | `signer:certificates:read`  | Reports whether the signer answers its health route, and how fast |

Query strings, request bodies and the signer's response pass through unchanged. The certificate list is keyset-paged:
pass `limit` and the last `next_cursor` as `cursor`. A signer that cannot be reached returns 502.

## Configuration

Settings are in the `connectors.signer` block of [`config/custos.yaml`](../../config/custos.yaml). With an empty
`base_url` the connector skips loading and reports the gap on `GET /connectors`.

| Key               | Variable in `config/custos.yaml` | Required | Purpose                       |
|-------------------|----------------------------------|----------|-------------------------------|
| `base_url`        | `SIGNER_BASE_URL`                | yes      | The signer's URL              |
| `timeout_seconds` |                                  | no       | Request timeout, default `10` |

The connector holds no signer client credentials. Those stay with the applications that call `POST /api/v1/sign`.
