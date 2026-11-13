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

The Signer connector is the only caller of
the [SSH certificate signer](../../extensions/SSH-Certificate-Signer/README.md).
It decides who may have a certificate, then calls the signer with one service token. The caller's own token never goes
to the signer. To offer the signer without the rest of Custos, enable only this connector.

## Routes

| Route                                                     | Needs                       | Does                                                      |
|-----------------------------------------------------------|-----------------------------|-----------------------------------------------------------|
| `GET /connectors/signer/certificates`                     | `signer:certificates:read`  | All certificates                                          |
| `GET /connectors/signer/certificates/{serial}`            | `signer:certificates:read`  | Any certificate                                           |
| `POST /connectors/signer/certificates/{serial}/revoke`    | `signer:certificates:write` | Revoke any certificate, recorded as `user/<admin id>`     |
| `GET /connectors/signer/me/certificates`                  | Signed in                   | The caller's own certificates                             |
| `POST /connectors/signer/me/certificates`                 | Signed in                   | Generate a certificate for the caller's account           |
| `GET /connectors/signer/me/certificates/{serial}`         | Signed in                   | One of the caller's own certificates                      |
| `POST /connectors/signer/me/certificates/{serial}/revoke` | Signed in                   | Revoke one of the caller's own, recorded as `user/<id>`   |
| `GET /connectors/signer/status`                           | `signer:certificates:read`  | Whether the signer answers its health route, and how fast |

The certificate lists are keyset-paged: pass `limit` and the last `next_cursor` as `cursor`. Someone else's certificate
on a `me` route is a 404.

To generate, the caller needs an approved and provisioned cluster account on `cluster_id`, and an active membership in
an
active allocation on that cluster. The username comes from the cluster account, and the email and subject from the
caller's verified token, never from the request.

Besides `public_key`, the caller may send options that only narrow the certificate:

| Field                | Effect                                                                                     |
|----------------------|--------------------------------------------------------------------------------------------|
| `ttl_seconds`        | Shorter lifetime. Empty means `certificate_ttl_seconds`, which is also the longest allowed |
| `source_address`     | Comma-separated IP addresses or CIDR blocks the certificate works from                     |
| `force_command`      | The only command a login runs                                                              |
| `exclude_extensions` | Standard options to leave out, such as `permit-port-forwarding`                            |

A refusal has an `error` code: `user_not_active`, `no_cluster_account`, `cluster_account_not_ready`,
`no_active_allocation`, `invalid_ttl`, `signer_unavailable`, or from the signer `reserved_username`,
`invalid_public_key`, `key_type_not_allowed`, `invalid_request` and `too_many_certificates`.

Service accounts can't use these routes. A person generates a certificate and hands it to the tool that needs it.

## Configuration

Settings are in the `connectors.signer` block of [`config/custos.yaml`](../../config/custos.yaml). When a required
setting is empty the connector skips loading and reports the gap on `GET /connectors`.

| Key                       | Variable in `config/custos.yaml` | Required | Purpose                                        |
|---------------------------|----------------------------------|----------|------------------------------------------------|
| `base_url`                | `SIGNER_BASE_URL`                | yes      | The signer's URL. Keep it off public networks. |
| `service_token`           | `SIGNER_SERVICE_TOKEN`           | yes      | The same token set on the signer               |
| `cluster_id`              | `CUSTOS_CLUSTER_ID`              | yes      | The cluster certificates log in to             |
| `certificate_ttl_seconds` |                                  | no       | Longest certificate lifetime, default `14400`  |
| `timeout_seconds`         |                                  | no       | Request timeout, default `10`                  |
